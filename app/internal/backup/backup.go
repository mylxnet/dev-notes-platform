// Package backup 负责把数据库与上传图片打成一份自包含的备份包，并在需要时恢复回来。
//
// 包结构固定为：
//
//	manifest.json      清单：格式版本、程序版本、结构版本、逐文件 sha256、内容统计
//	notes.db           SQLite 一致性快照（VACUUM INTO 导出，含 WAL 中未落盘的数据）
//	media/<hash>.<ext> 后台上传的图片，文件名即内容哈希，恢复后链接无需改动
//
// 恢复的关键约束：替换数据库文件前必须先关闭连接并删除 -wal/-shm，
// 否则新库会和旧 WAL 拼在一起导致数据损坏。
package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	// 预检时直接用 sql 打开包内数据库，需要注册驱动。
	_ "modernc.org/sqlite"
)

const (
	// FormatVersion 标记备份包格式，将来结构变化时用它区分。
	FormatVersion = "dev-notes-backup/1"
	// ManifestName 是包内清单文件名。
	ManifestName = "manifest.json"
	// DatabaseName 是包内数据库文件名。
	DatabaseName = "notes.db"
	// MediaDirName 是包内图片目录名。
	MediaDirName = "media"
	// MaxArchiveBytes 是允许上传的备份包大小上限。
	MaxArchiveBytes = 500 << 20
	// TmpDirName 是数据目录下的临时工作目录。
	TmpDirName = ".tmp"
	// SnapshotsDirName 是恢复前快照目录。
	SnapshotsDirName = "backups"
)

// 备份包只接受这三类条目，其余一律拒绝，避免解包时写出意外文件。
var (
	uploadedNamePattern = regexp.MustCompile(`^[0-9a-f]{32}\.(?:png|jpg|jpeg|webp|gif)$`)
	errUnsafeEntry      = errors.New("备份包包含非法条目")
)

// Entry 记录包内单个文件的校验信息。
type Entry struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

// Counts 是内容统计，恢复后展示给用户看。
type Counts struct {
	Notes      int   `json:"notes"`
	Projects   int   `json:"projects"`
	MediaFiles int   `json:"media_files"`
	MediaBytes int64 `json:"media_bytes"`
}

// Manifest 是备份包的清单。
type Manifest struct {
	Format        string    `json:"format"`
	AppVersion    string    `json:"app_version"`
	SchemaVersion int       `json:"schema_version"`
	CreatedAt     time.Time `json:"created_at"`
	Entries       []Entry   `json:"entries"`
	Counts        Counts    `json:"counts"`
}

// Snapshotter 由 store.Store 实现；抽成接口便于测试替换。
type Snapshotter interface {
	SnapshotTo(ctx context.Context, dest string) error
	SchemaVersionOf(ctx context.Context) (int, error)
	ContentCounts(ctx context.Context) (notes int, projects int, err error)
	MediaReferences(ctx context.Context) ([]string, error)
}

// Inspection 是预检结果：能不能恢复、有什么问题。
type Inspection struct {
	Manifest   Manifest
	ExtractDir string
	// Missing 是数据库引用但包里没有的图片。
	Missing []string
	// Problems 阻断恢复的问题；Warnings 只是提醒。
	Problems []string
	Warnings []string
	// Restorable 为 true 才允许执行恢复。
	Restorable bool
}

// Create 生成备份包到 outPath，返回写入清单的信息。
func Create(ctx context.Context, src Snapshotter, dataDir, outPath, appVersion string) (*Manifest, error) {
	if err := EnsureDirs(dataDir); err != nil {
		return nil, err
	}
	workDir, err := os.MkdirTemp(filepath.Join(dataDir, TmpDirName), "backup-*")
	if err != nil {
		return nil, fmt.Errorf("create backup workdir: %w", err)
	}
	defer os.RemoveAll(workDir)

	snapPath := filepath.Join(workDir, DatabaseName)
	if err := src.SnapshotTo(ctx, snapPath); err != nil {
		return nil, err
	}

	// 清单里的统计取自快照本身，保证与包内容一致。
	info, err := inspectDatabase(snapPath)
	if err != nil {
		return nil, err
	}

	mediaDir := filepath.Join(dataDir, MediaDirName)
	mediaNames, err := listUploadedMedia(mediaDir)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(mediaNames)+1)
	dbEntry, err := fileEntry(DatabaseName, snapPath)
	if err != nil {
		return nil, err
	}
	entries = append(entries, dbEntry)

	var mediaBytes int64
	for _, name := range mediaNames {
		entry, err := fileEntry(MediaDirName+"/"+name, filepath.Join(mediaDir, name))
		if err != nil {
			return nil, err
		}
		mediaBytes += entry.Bytes
		entries = append(entries, entry)
	}

	manifest := Manifest{
		Format:        FormatVersion,
		AppVersion:    appVersion,
		SchemaVersion: info.schemaVersion,
		CreatedAt:     time.Now().UTC(),
		Entries:       entries,
		Counts: Counts{
			Notes:      info.notes,
			Projects:   info.projects,
			MediaFiles: len(mediaNames),
			MediaBytes: mediaBytes,
		},
	}

	if err := writeArchive(outPath, manifest, snapPath, mediaDir, mediaNames); err != nil {
		_ = os.Remove(outPath)
		return nil, err
	}
	return &manifest, nil
}

// Inspect 解包并逐项校验；currentSchema 用于判断版本兼容性。
// 解包目录由调用方负责清理。
func Inspect(archivePath, workDir string, currentSchema int) (*Inspection, error) {
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return nil, fmt.Errorf("create extract dir: %w", err)
	}
	hashes, err := extractArchive(archivePath, workDir)
	if err != nil {
		return nil, err
	}

	inspection := &Inspection{ExtractDir: workDir}

	if _, ok := hashes[ManifestName]; !ok {
		inspection.Problems = append(inspection.Problems, "备份包里没有 manifest.json，无法确认内容")
		return inspection, nil
	}
	manifest, err := readManifest(filepath.Join(workDir, ManifestName))
	if err != nil {
		inspection.Problems = append(inspection.Problems, "manifest.json 解析失败："+err.Error())
		return inspection, nil
	}
	inspection.Manifest = manifest

	if manifest.Format != FormatVersion {
		inspection.Problems = append(inspection.Problems,
			fmt.Sprintf("备份包格式为 %s，当前程序只支持 %s", manifest.Format, FormatVersion))
	}
	if _, ok := hashes[DatabaseName]; !ok {
		inspection.Problems = append(inspection.Problems, "备份包里缺少 notes.db")
	} else if _, err := os.Stat(filepath.Join(workDir, DatabaseName)); err != nil {
		inspection.Problems = append(inspection.Problems, "notes.db 解包失败")
	}

	// 逐文件核对 sha256：manifest 里有的必须存在且一致。
	declared := make(map[string]string, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		declared[entry.Path] = entry.SHA256
	}
	for entryPath, want := range declared {
		got, ok := hashes[entryPath]
		if !ok {
			inspection.Problems = append(inspection.Problems, "备份包缺少条目："+entryPath)
			continue
		}
		if got != want {
			inspection.Problems = append(inspection.Problems, "文件校验失败："+entryPath)
		}
	}
	for entryPath := range hashes {
		if entryPath == ManifestName {
			continue
		}
		if _, ok := declared[entryPath]; !ok {
			inspection.Problems = append(inspection.Problems, "备份包包含清单外的条目："+entryPath)
		}
	}

	if manifest.SchemaVersion > currentSchema {
		inspection.Problems = append(inspection.Problems, fmt.Sprintf(
			"备份来自更高版本（结构版本 %d，当前程序只到 %d），请先升级程序再恢复",
			manifest.SchemaVersion, currentSchema))
	}

	if _, ok := hashes[DatabaseName]; ok {
		dbPath := filepath.Join(workDir, DatabaseName)
		if info, err := inspectDatabase(dbPath); err != nil {
			inspection.Problems = append(inspection.Problems, "数据库不可用："+err.Error())
		} else {
			if info.schemaVersion > currentSchema {
				inspection.Problems = append(inspection.Problems,
					fmt.Sprintf("数据库结构版本 %d 高于当前程序 %d", info.schemaVersion, currentSchema))
			}
			if info.notes != manifest.Counts.Notes || info.projects != manifest.Counts.Projects {
				inspection.Warnings = append(inspection.Warnings, fmt.Sprintf(
					"清单统计与实际内容不一致（清单 %d 篇 / %d 个项目，实际 %d 篇 / %d 个项目）",
					manifest.Counts.Notes, manifest.Counts.Projects, info.notes, info.projects))
			}
			inspection.Warnings = append(inspection.Warnings,
				fmt.Sprintf("数据库完整性检查通过，共 %d 篇笔记、%d 个项目、结构版本 %d",
					info.notes, info.projects, info.schemaVersion))
		}

		present := make(map[string]struct{}, len(hashes))
		for entryPath := range hashes {
			if strings.HasPrefix(entryPath, MediaDirName+"/") {
				present[strings.TrimPrefix(entryPath, MediaDirName+"/")] = struct{}{}
			}
		}
		refs, err := databaseMediaReferences(dbPath)
		if err != nil {
			inspection.Warnings = append(inspection.Warnings, "无法读取图片引用："+err.Error())
		} else {
			for _, ref := range refs {
				if _, ok := present[ref]; !ok {
					inspection.Missing = append(inspection.Missing, ref)
				}
			}
			sort.Strings(inspection.Missing)
			if len(inspection.Missing) > 0 {
				inspection.Warnings = append(inspection.Warnings, fmt.Sprintf(
					"有 %d 张图片被内容引用但不在包里，恢复后这些图会显示为破图", len(inspection.Missing)))
			} else if len(refs) > 0 {
				inspection.Warnings = append(inspection.Warnings,
					fmt.Sprintf("内容引用的 %d 张图片在包内全部存在", len(refs)))
			}
		}
	}

	inspection.Restorable = len(inspection.Problems) == 0
	return inspection, nil
}

// Apply 用解包目录覆盖数据目录里的数据库与图片。
// 调用前必须已关闭数据库连接；失败时会尝试回滚数据库文件。
func Apply(ctx context.Context, extractDir, dataDir, dbPath string) error {
	if _, err := os.Stat(filepath.Join(extractDir, DatabaseName)); err != nil {
		return fmt.Errorf("备份包缺少数据库文件: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return fmt.Errorf("prepare db dir: %w", err)
	}

	// 旧库先留一份，失败时还原。
	rollbackPath := filepath.Join(dataDir, TmpDirName, DatabaseName+".rollback")
	if err := os.MkdirAll(filepath.Dir(rollbackPath), 0o700); err != nil {
		return fmt.Errorf("prepare rollback dir: %w", err)
	}
	_ = os.Remove(rollbackPath)
	if _, err := os.Stat(dbPath); err == nil {
		if err := copyFile(dbPath, rollbackPath); err != nil {
			return fmt.Errorf("backup current database: %w", err)
		}
	}

	restore := func() {
		if _, err := os.Stat(rollbackPath); err == nil {
			_ = copyFile(rollbackPath, dbPath)
		}
		_ = os.Remove(dbPath + "-wal")
		_ = os.Remove(dbPath + "-shm")
	}

	// 关键：先删掉旧库的 WAL/SHM，否则新库会与旧日志混在一起。
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := os.Remove(dbPath + suffix); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove stale %s: %w", suffix, err)
		}
	}
	if err := copyFile(filepath.Join(extractDir, DatabaseName), dbPath); err != nil {
		restore()
		return fmt.Errorf("write database: %w", err)
	}

	mediaDir := filepath.Join(dataDir, MediaDirName)
	if err := os.MkdirAll(mediaDir, 0o700); err != nil {
		restore()
		return fmt.Errorf("prepare media dir: %w", err)
	}
	entries, err := os.ReadDir(filepath.Join(extractDir, MediaDirName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		restore()
		return fmt.Errorf("read backup media: %w", err)
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			restore()
			return fmt.Errorf("恢复被中断: %w", err)
		}
		if entry.IsDir() || !uploadedNamePattern.MatchString(entry.Name()) {
			continue
		}
		source := filepath.Join(extractDir, MediaDirName, entry.Name())
		if err := copyFile(source, filepath.Join(mediaDir, entry.Name())); err != nil {
			restore()
			return fmt.Errorf("write media %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// DiscardRollback 在恢复成功（新库已能打开）后丢弃回滚文件。
func DiscardRollback(dataDir string) {
	_ = os.Remove(filepath.Join(dataDir, TmpDirName, DatabaseName+".rollback"))
}

// Rollback 用回滚文件还原上一份数据库。
func Rollback(dataDir, dbPath string) error {
	rollbackPath := filepath.Join(dataDir, TmpDirName, DatabaseName+".rollback")
	if _, err := os.Stat(rollbackPath); err != nil {
		return fmt.Errorf("没有可用的回滚文件: %w", err)
	}
	if err := copyFile(rollbackPath, dbPath); err != nil {
		return fmt.Errorf("rollback database: %w", err)
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		_ = os.Remove(dbPath + suffix)
	}
	return nil
}

// EnsureDirs 预先建好数据目录下的临时目录与快照目录。
func EnsureDirs(dataDir string) error {
	for _, dir := range []string{TmpDirName, SnapshotsDirName} {
		if err := os.MkdirAll(filepath.Join(dataDir, dir), 0o700); err != nil {
			return fmt.Errorf("create %s dir: %w", dir, err)
		}
	}
	return nil
}

// CleanTmp 清理数据目录下的临时文件（启动时与备份结束后调用）。
func CleanTmp(dataDir string) {
	entries, err := os.ReadDir(filepath.Join(dataDir, TmpDirName))
	if err != nil {
		return
	}
	for _, entry := range entries {
		_ = os.RemoveAll(filepath.Join(dataDir, TmpDirName, entry.Name()))
	}
}

// SnapshotPath 生成恢复前快照的路径。
func SnapshotPath(dataDir string, now time.Time) string {
	return filepath.Join(dataDir, SnapshotsDirName,
		"pre-restore-"+now.UTC().Format("20060102-150405")+".tar.gz")
}

// MediaStat 统计数据目录里上传图片的数量与占用。
func MediaStat(dataDir string) (count int, bytes int64, err error) {
	names, err := listUploadedMedia(filepath.Join(dataDir, MediaDirName))
	if err != nil {
		return 0, 0, err
	}
	for _, name := range names {
		info, err := os.Stat(filepath.Join(dataDir, MediaDirName, name))
		if err != nil {
			continue
		}
		bytes += info.Size()
	}
	return len(names), bytes, nil
}

// MissingMedia 返回内容引用到、但数据目录里不存在的图片文件名。
func MissingMedia(dataDir string, refs []string) []string {
	present := make(map[string]struct{}, len(refs))
	names, err := listUploadedMedia(filepath.Join(dataDir, MediaDirName))
	if err == nil {
		for _, name := range names {
			present[name] = struct{}{}
		}
	}
	missing := make([]string, 0)
	for _, ref := range refs {
		if _, ok := present[ref]; !ok {
			missing = append(missing, ref)
		}
	}
	sort.Strings(missing)
	return missing
}

// ListSnapshots 列出数据目录里的恢复前快照（新的在前）。
type Snapshot struct {
	Name    string    `json:"name"`
	Bytes   int64     `json:"bytes"`
	ModTime time.Time `json:"mod_time"`
}

// ListSnapshots 返回 backups/ 下的快照文件。
func ListSnapshots(dataDir string) ([]Snapshot, error) {
	entries, err := os.ReadDir(filepath.Join(dataDir, SnapshotsDirName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Snapshot{}, nil
		}
		return nil, fmt.Errorf("read snapshots: %w", err)
	}
	items := make([]Snapshot, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, Snapshot{Name: entry.Name(), Bytes: info.Size(), ModTime: info.ModTime().UTC()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ModTime.After(items[j].ModTime) })
	return items, nil
}

// ---------- 内部实现 ----------

type databaseInfo struct {
	schemaVersion int
	notes         int
	projects      int
}

// inspectDatabase 用只读方式确认库文件完好，并读出结构版本与内容统计。
func inspectDatabase(dbPath string) (databaseInfo, error) {
	var info databaseInfo
	if _, err := os.Stat(dbPath); err != nil {
		return info, fmt.Errorf("database file missing: %w", err)
	}
	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return info, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	var check string
	if err := db.QueryRow("PRAGMA integrity_check").Scan(&check); err != nil {
		return info, fmt.Errorf("integrity check: %w", err)
	}
	if check != "ok" {
		return info, fmt.Errorf("integrity check failed: %s", check)
	}
	if err := db.QueryRow("PRAGMA user_version").Scan(&info.schemaVersion); err != nil {
		return info, fmt.Errorf("read user_version: %w", err)
	}
	if err := db.QueryRow(
		`SELECT (SELECT COUNT(*) FROM note), (SELECT COUNT(*) FROM project)`).
		Scan(&info.notes, &info.projects); err != nil {
		return info, fmt.Errorf("count content: %w", err)
	}
	return info, nil
}

// databaseMediaReferences 从备份包里的库读出图片引用，不依赖 store 包。
func databaseMediaReferences(dbPath string) ([]string, error) {
	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT content FROM note
		UNION ALL SELECT description FROM project
		UNION ALL SELECT cover_image FROM project`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		for _, match := range mediaRefInText(text) {
			seen[match] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	refs := make([]string, 0, len(seen))
	for name := range seen {
		refs = append(refs, name)
	}
	sort.Strings(refs)
	return refs, nil
}

func fileEntry(archivePath, diskPath string) (Entry, error) {
	info, err := os.Stat(diskPath)
	if err != nil {
		return Entry{}, fmt.Errorf("stat %s: %w", archivePath, err)
	}
	sum, err := fileSHA256(diskPath)
	if err != nil {
		return Entry{}, err
	}
	return Entry{Path: archivePath, Bytes: info.Size(), SHA256: sum}, nil
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func listUploadedMedia(mediaDir string) ([]string, error) {
	entries, err := os.ReadDir(mediaDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read media dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !uploadedNamePattern.MatchString(entry.Name()) {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func writeArchive(outPath string, manifest Manifest, dbSnap, mediaDir string, mediaNames []string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o700); err != nil {
		return fmt.Errorf("prepare archive dir: %w", err)
	}
	file, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewWriterLevel(file, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("gzip writer: %w", err)
	}
	gz.ModTime = manifest.CreatedAt
	tw := tar.NewWriter(gz)

	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	if err := writeTarBytes(tw, ManifestName, raw, manifest.CreatedAt); err != nil {
		return err
	}
	if err := writeTarFile(tw, DatabaseName, dbSnap, manifest.CreatedAt); err != nil {
		return err
	}
	for _, name := range mediaNames {
		if err := writeTarFile(tw, MediaDirName+"/"+name, filepath.Join(mediaDir, name), manifest.CreatedAt); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("close tar: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("close gzip: %w", err)
	}
	return nil
}

func writeTarBytes(tw *tar.Writer, name string, data []byte, modTime time.Time) error {
	header := &tar.Header{
		Name:     name,
		Mode:     0o600,
		Size:     int64(len(data)),
		Typeflag: tar.TypeReg,
		ModTime:  modTime,
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write header %s: %w", name, err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

func writeTarFile(tw *tar.Writer, name, diskPath string, modTime time.Time) error {
	info, err := os.Stat(diskPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", diskPath, err)
	}
	header := &tar.Header{
		Name:     name,
		Mode:     0o600,
		Size:     info.Size(),
		Typeflag: tar.TypeReg,
		ModTime:  modTime,
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write header %s: %w", name, err)
	}
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", diskPath, err)
	}
	defer file.Close()
	if _, err := io.Copy(tw, file); err != nil {
		return fmt.Errorf("copy %s: %w", diskPath, err)
	}
	return nil
}

// extractArchive 解包到 destDir，返回「条目名 → sha256」。
// 只允许 manifest.json、notes.db 与 media/<合法文件名>，拒绝路径穿越与绝对路径。
func extractArchive(archivePath, destDir string) (map[string]string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("不是有效的 tar.gz 备份包: %w", err)
	}
	defer gz.Close()

	hashes := make(map[string]string)
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取备份包失败: %w", err)
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if header.Typeflag != tar.TypeReg {
			return nil, fmt.Errorf("%w: 不支持的文件类型 %s", errUnsafeEntry, header.Name)
		}
		clean, err := safeEntryName(header.Name)
		if err != nil {
			return nil, err
		}
		target := filepath.Join(destDir, filepath.FromSlash(clean))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return nil, fmt.Errorf("prepare extract dir: %w", err)
		}
		hash := sha256.New()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", clean, err)
		}
		if _, err := io.Copy(io.MultiWriter(out, hash), tr); err != nil {
			out.Close()
			return nil, fmt.Errorf("extract %s: %w", clean, err)
		}
		if err := out.Close(); err != nil {
			return nil, fmt.Errorf("close %s: %w", clean, err)
		}
		hashes[clean] = hex.EncodeToString(hash.Sum(nil))
	}
	return hashes, nil
}

// safeEntryName 归一化并校验条目路径。
func safeEntryName(name string) (string, error) {
	clean := path.Clean(strings.TrimPrefix(name, "./"))
	if clean == "." || path.IsAbs(clean) || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("%w: %s", errUnsafeEntry, name)
	}
	switch {
	case clean == ManifestName, clean == DatabaseName:
		return clean, nil
	case strings.HasPrefix(clean, MediaDirName+"/"):
		base := strings.TrimPrefix(clean, MediaDirName+"/")
		if strings.Contains(base, "/") || !uploadedNamePattern.MatchString(base) {
			return "", fmt.Errorf("%w: %s", errUnsafeEntry, name)
		}
		return clean, nil
	default:
		return "", fmt.Errorf("%w: %s", errUnsafeEntry, name)
	}
}

func readManifest(path string) (Manifest, error) {
	var manifest Manifest
	raw, err := os.ReadFile(path)
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func copyFile(source, target string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	tmp := target + ".part"
	dst, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// mediaRefInText 从一段文本里提取上传图片文件名。
func mediaRefInText(text string) []string {
	matches := mediaRefPattern.FindAllStringSubmatch(text, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		names = append(names, match[1])
	}
	return names
}

var mediaRefPattern = regexp.MustCompile(`/media/uploads/([0-9a-f]{32}\.(?:png|jpg|jpeg|webp|gif))`)
