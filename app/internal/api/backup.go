package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/backup"
	"dev-notes-platform/internal/store"
)

// maxBackupUpload 比包上限略大，留给 multipart 头部开销。
const maxBackupUpload = backup.MaxArchiveBytes + (8 << 20)

// snapshotNamePattern 限定可回滚的快照文件名，杜绝路径拼接穿越。
var snapshotNamePattern = regexp.MustCompile(`^pre-restore-[0-9]{8}-[0-9]{6}\.tar\.gz$`)

// backupSummary 返回备份页需要的规模信息与已有快照。
func (s *Server) adminBackupSummary(c *gin.Context) {
	ctx := c.Request.Context()

	var dbBytes int64
	if info, err := os.Stat(s.dbPath); err == nil {
		dbBytes = info.Size()
	}
	mediaFiles, mediaBytes, err := backup.MediaStat(s.dataDir)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "统计图片失败："+err.Error())
		return
	}
	notes, projects, err := s.store().ContentCounts(ctx)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "统计内容失败："+err.Error())
		return
	}
	schema, err := s.store().SchemaVersionOf(ctx)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "读取数据库版本失败："+err.Error())
		return
	}
	snapshots, err := backup.ListSnapshots(s.dataDir)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "读取快照列表失败："+err.Error())
		return
	}

	respondOK(c, gin.H{
		"db_bytes":          dbBytes,
		"media_files":       mediaFiles,
		"media_bytes":       mediaBytes,
		"notes":             notes,
		"projects":          projects,
		"schema_version":    schema,
		"app_version":       s.version,
		"snapshots":         snapshots,
		"max_archive_bytes": backup.MaxArchiveBytes,
	})
}

// adminCreateBackup 生成备份包并直接以附件下发；服务器上只留临时文件，下载后即删。
func (s *Server) adminCreateBackup(c *gin.Context) {
	if err := backup.EnsureDirs(s.dataDir); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "准备目录失败："+err.Error())
		return
	}
	now := time.Now().UTC()
	target := filepath.Join(s.dataDir, backup.TmpDirName, "download-"+now.Format("20060102-150405")+".tar.gz")

	manifest, err := backup.Create(c.Request.Context(), s.store(), s.dataDir, target, s.version)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "BACKUP_FAILED", "生成备份失败："+err.Error())
		return
	}
	defer os.Remove(target)

	c.Header("X-Backup-Notes", fmt.Sprint(manifest.Counts.Notes))
	c.Header("X-Backup-Media", fmt.Sprint(manifest.Counts.MediaFiles))
	c.FileAttachment(target, "dev-notes-backup-"+now.Format("20060102-1504")+".tar.gz")
}

// adminInspectBackup 只做预检：解包、校验 sha256、判断版本与图片是否齐全，不改动任何数据。
func (s *Server) adminInspectBackup(c *gin.Context) {
	upload, err := s.saveBackupUpload(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	defer os.Remove(upload)

	workDir := filepath.Join(s.dataDir, backup.TmpDirName, "inspect-"+fmt.Sprint(time.Now().UnixNano()))
	defer os.RemoveAll(workDir)

	inspection, err := backup.Inspect(upload, workDir, store.CurrentSchemaVersion())
	if err != nil {
		respondError(c, http.StatusUnprocessableEntity, "BACKUP_INVALID", "备份包不可用："+err.Error())
		return
	}

	missing := inspection.Missing
	if missing == nil {
		missing = []string{}
	}
	problems := inspection.Problems
	if problems == nil {
		problems = []string{}
	}
	warnings := inspection.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	respondOK(c, gin.H{
		"restorable": inspection.Restorable,
		"manifest":   inspection.Manifest,
		"missing":    missing,
		"problems":   problems,
		"warnings":   warnings,
	})
}

// adminRestoreBackup 用上传的包覆盖当前数据；执行前会先给当前数据留一份快照。
func (s *Server) adminRestoreBackup(c *gin.Context) {
	if strings.ToUpper(strings.TrimSpace(c.PostForm("confirm"))) != "RESTORE" {
		respondError(c, http.StatusBadRequest, "CONFIRM_REQUIRED", "请输入 RESTORE 确认覆盖当前数据")
		return
	}
	if !s.restoreMu.TryLock() {
		respondError(c, http.StatusConflict, "RESTORE_BUSY", "已有恢复任务在执行，请稍候")
		return
	}
	defer s.restoreMu.Unlock()

	upload, err := s.saveBackupUpload(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	defer os.Remove(upload)

	s.maintenance.Store(true)
	defer s.maintenance.Store(false)

	outcome, failure := s.applyArchive(c.Request.Context(), upload)
	if failure != nil {
		respondError(c, failure.status, failure.code, failure.message)
		return
	}
	respondOK(c, outcome)
}

// adminListSnapshots 列出恢复前留下的快照，供一键回滚。
func (s *Server) adminListSnapshots(c *gin.Context) {
	snapshots, err := backup.ListSnapshots(s.dataDir)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "读取快照失败："+err.Error())
		return
	}
	respondOK(c, gin.H{"items": snapshots})
}

// adminRestoreSnapshot 用服务器上的恢复前快照回滚，逻辑与上传恢复一致。
func (s *Server) adminRestoreSnapshot(c *gin.Context) {
	name := c.Param("name")
	if !snapshotNamePattern.MatchString(name) {
		respondError(c, http.StatusBadRequest, "INVALID_NAME", "快照文件名不合法")
		return
	}
	archive := filepath.Join(s.dataDir, backup.SnapshotsDirName, name)
	if _, err := os.Stat(archive); err != nil {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "快照不存在")
		return
	}
	if !s.restoreMu.TryLock() {
		respondError(c, http.StatusConflict, "RESTORE_BUSY", "已有恢复任务在执行，请稍候")
		return
	}
	defer s.restoreMu.Unlock()

	s.maintenance.Store(true)
	defer s.maintenance.Store(false)

	outcome, failure := s.applyArchive(c.Request.Context(), archive)
	if failure != nil {
		respondError(c, failure.status, failure.code, failure.message)
		return
	}
	respondOK(c, outcome)
}

// restoreOutcome 是恢复成功后的自检结果。
type restoreOutcome struct {
	Notes         int      `json:"notes"`
	Projects      int      `json:"projects"`
	MediaFiles    int      `json:"media_files"`
	Missing       []string `json:"missing"`
	Snapshot      string   `json:"snapshot"`
	SchemaVersion int      `json:"schema_version"`
}

// restoreFailure 携带要回给前端的状态码与提示。
type restoreFailure struct {
	status  int
	code    string
	message string
}

// applyArchive 是恢复的核心流程：预检 → 快照当前数据 → 关库 → 覆盖文件 → 重开 → 自检。
// 任何一步失败都会把数据库还原到恢复前的状态。
func (s *Server) applyArchive(ctx context.Context, archivePath string) (*restoreOutcome, *restoreFailure) {
	workDir := filepath.Join(s.dataDir, backup.TmpDirName, "restore-"+fmt.Sprint(time.Now().UnixNano()))
	defer os.RemoveAll(workDir)

	inspection, err := backup.Inspect(archivePath, workDir, store.CurrentSchemaVersion())
	if err != nil {
		return nil, &restoreFailure{
			status:  http.StatusUnprocessableEntity,
			code:    "BACKUP_INVALID",
			message: "备份包不可用：" + err.Error(),
		}
	}
	if !inspection.Restorable {
		return nil, &restoreFailure{
			status:  http.StatusUnprocessableEntity,
			code:    "BACKUP_INCOMPATIBLE",
			message: "备份包不能恢复到当前版本：" + strings.Join(inspection.Problems, "；"),
		}
	}

	if err := backup.EnsureDirs(s.dataDir); err != nil {
		return nil, &restoreFailure{http.StatusInternalServerError, "INTERNAL_ERROR", "准备目录失败：" + err.Error()}
	}

	// 覆盖不可逆，先把当前数据打成快照；后面任何一步失败都能靠它回滚。
	snapshotPath := backup.SnapshotPath(s.dataDir, time.Now())
	if _, err := backup.Create(ctx, s.store(), s.dataDir, snapshotPath, s.version); err != nil {
		return nil, &restoreFailure{
			status:  http.StatusInternalServerError,
			code:    "SNAPSHOT_FAILED",
			message: "恢复前快照失败，已取消恢复：" + err.Error(),
		}
	}

	old := s.store()
	if err := old.Close(); err != nil {
		return nil, &restoreFailure{http.StatusInternalServerError, "RESTORE_FAILED", "关闭数据库失败：" + err.Error()}
	}

	// 重开失败时至少让服务带着旧库继续跑，避免整站 500。
	reopenOld := func() {
		if reopened, err := store.Open(context.Background(), s.dbPath); err == nil {
			s.db.swap(reopened)
		}
	}

	if err := backup.Apply(ctx, workDir, s.dataDir, s.dbPath); err != nil {
		_ = backup.Rollback(s.dataDir, s.dbPath)
		reopenOld()
		return nil, &restoreFailure{
			status:  http.StatusInternalServerError,
			code:    "RESTORE_FAILED",
			message: "恢复失败，已回滚到恢复前的数据：" + err.Error(),
		}
	}

	next, err := store.Open(ctx, s.dbPath)
	if err != nil {
		_ = backup.Rollback(s.dataDir, s.dbPath)
		reopenOld()
		return nil, &restoreFailure{
			status:  http.StatusInternalServerError,
			code:    "RESTORE_FAILED",
			message: "恢复后的数据库无法打开，已回滚到恢复前的数据：" + err.Error(),
		}
	}
	s.db.swap(next)
	backup.DiscardRollback(s.dataDir)

	// 自检用独立上下文：恢复已成功，不该因为原请求断开而报错。
	check := context.Background()
	notes, projects, _ := next.ContentCounts(check)
	schema, _ := next.SchemaVersionOf(check)
	mediaFiles, _, _ := backup.MediaStat(s.dataDir)
	missing := []string{}
	if refs, err := next.MediaReferences(check); err == nil {
		missing = backup.MissingMedia(s.dataDir, refs)
	}

	return &restoreOutcome{
		Notes:         notes,
		Projects:      projects,
		MediaFiles:    mediaFiles,
		Missing:       missing,
		Snapshot:      filepath.Base(snapshotPath),
		SchemaVersion: schema,
	}, nil
}

// saveBackupUpload 把上传的备份包流式落到临时目录，返回本地路径。
func (s *Server) saveBackupUpload(c *gin.Context) (string, error) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return "", errors.New("请选择要上传的备份包")
	}
	defer file.Close()

	if header.Size > maxBackupUpload {
		return "", fmt.Errorf("备份包不能超过 %d MB", backup.MaxArchiveBytes>>20)
	}
	if !strings.HasSuffix(strings.ToLower(filepath.Base(header.Filename)), ".tar.gz") {
		return "", errors.New("只接受 .tar.gz 备份包")
	}

	dir := filepath.Join(s.dataDir, backup.TmpDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("创建临时目录失败：%w", err)
	}
	target := filepath.Join(dir, "upload-"+fmt.Sprint(time.Now().UnixNano())+".tar.gz")
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", fmt.Errorf("保存上传文件失败：%w", err)
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		_ = os.Remove(target)
		return "", fmt.Errorf("保存上传文件失败：%w", err)
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(target)
		return "", fmt.Errorf("保存上传文件失败：%w", err)
	}
	return target, nil
}
