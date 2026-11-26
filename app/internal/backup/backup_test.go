package backup_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dev-notes-platform/internal/backup"
	"dev-notes-platform/internal/model"
	"dev-notes-platform/internal/store"
)

// fixture 造一份「一个项目 + 一篇笔记 + 一张上传图」的数据目录。
type fixture struct {
	dataDir  string
	dbPath   string
	mediaDir string
	media    string
	st       *store.Store
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "notes.db")

	st, err := store.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	mediaDir := filepath.Join(dataDir, "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		t.Fatalf("create media dir: %v", err)
	}
	name := strings.Repeat("a", 32) + ".png"
	if err := os.WriteFile(filepath.Join(mediaDir, name), []byte("fake-png-bytes"), 0o644); err != nil {
		t.Fatalf("write media: %v", err)
	}

	project := &model.Project{
		Name:       "带封面的项目",
		Slug:       "with-cover",
		Summary:    "摘要",
		RepoURL:    "https://github.com/example/cover",
		CoverImage: "/media/uploads/" + name,
		Status:     model.StatusPublished,
	}
	if _, err := st.CreateProject(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}
	note := &model.Note{
		Title:     "带插图的笔记",
		Slug:      "with-image",
		Summary:   "摘要",
		Content:   "正文![](" + "/media/uploads/" + name + ")",
		ProjectID: &project.ID,
		Status:    model.StatusPublished,
	}
	if _, err := st.CreateNote(ctx, note); err != nil {
		t.Fatalf("create note: %v", err)
	}

	return &fixture{dataDir: dataDir, dbPath: dbPath, mediaDir: mediaDir, media: name, st: st}
}

func (f *fixture) mediaPath() string { return filepath.Join(f.mediaDir, f.media) }

// TestCreateAndInspect 验证包结构、清单统计与预检结论。
func TestCreateAndInspect(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	archive := filepath.Join(t.TempDir(), "backup.tar.gz")
	manifest, err := backup.Create(ctx, f.st, f.dataDir, archive, "1.5.0")
	if err != nil {
		t.Fatalf("create backup: %v", err)
	}
	if manifest.AppVersion != "1.5.0" {
		t.Fatalf("unexpected app version: %s", manifest.AppVersion)
	}
	if manifest.SchemaVersion != store.CurrentSchemaVersion() {
		t.Fatalf("expected schema %d, got %d", store.CurrentSchemaVersion(), manifest.SchemaVersion)
	}
	if manifest.Counts.Notes != 1 || manifest.Counts.Projects != 1 || manifest.Counts.MediaFiles != 1 {
		t.Fatalf("unexpected counts: %+v", manifest.Counts)
	}
	if len(manifest.Entries) != 2 {
		t.Fatalf("expected notes.db + 1 media entry, got %d", len(manifest.Entries))
	}

	inspection, err := backup.Inspect(archive, t.TempDir(), store.CurrentSchemaVersion())
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if !inspection.Restorable {
		t.Fatalf("expected restorable, problems: %v", inspection.Problems)
	}
	if len(inspection.Missing) != 0 {
		t.Fatalf("expected no missing media, got %v", inspection.Missing)
	}
	if inspection.Manifest.Counts.Notes != 1 {
		t.Fatalf("unexpected inspection manifest: %+v", inspection.Manifest)
	}
}

// TestInspectRejectsHigherSchema 验证备份来自更高结构版本时拒绝恢复。
func TestInspectRejectsHigherSchema(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	archive := filepath.Join(t.TempDir(), "backup.tar.gz")
	if _, err := backup.Create(ctx, f.st, f.dataDir, archive, "1.5.0"); err != nil {
		t.Fatalf("create backup: %v", err)
	}

	inspection, err := backup.Inspect(archive, t.TempDir(), store.CurrentSchemaVersion()-1)
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if inspection.Restorable {
		t.Fatalf("expected not restorable for higher schema, warnings: %v", inspection.Warnings)
	}
	joined := strings.Join(inspection.Problems, "；")
	if !strings.Contains(joined, "更高版本") {
		t.Fatalf("expected version problem, got %v", inspection.Problems)
	}
}

// TestInspectDetectsMissingMedia 验证内容引用了包里没有的图片时会报告出来（不阻断恢复）。
func TestInspectDetectsMissingMedia(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	if err := os.Remove(f.mediaPath()); err != nil {
		t.Fatalf("remove media: %v", err)
	}

	archive := filepath.Join(t.TempDir(), "backup.tar.gz")
	if _, err := backup.Create(ctx, f.st, f.dataDir, archive, "1.5.0"); err != nil {
		t.Fatalf("create backup: %v", err)
	}
	inspection, err := backup.Inspect(archive, t.TempDir(), store.CurrentSchemaVersion())
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if !inspection.Restorable {
		t.Fatalf("missing media should not block restore, problems: %v", inspection.Problems)
	}
	if len(inspection.Missing) != 1 || inspection.Missing[0] != f.media {
		t.Fatalf("expected missing %s, got %v", f.media, inspection.Missing)
	}
}

// TestInspectRejectsUnsafeEntries 验证路径穿越条目会被拒绝。
func TestInspectRejectsUnsafeEntries(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "bad.tar.gz")

	file, err := os.Create(archive)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	writeTarEntry(t, tw, "../evil.txt", []byte("x"))
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}

	if _, err := backup.Inspect(archive, filepath.Join(dir, "out"), store.CurrentSchemaVersion()); err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}

// TestApplyRestoresData 验证「备份 → 数据丢失 → 恢复」后内容与图片都回来了。
func TestApplyRestoresData(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	archive := filepath.Join(t.TempDir(), "backup.tar.gz")
	if _, err := backup.Create(ctx, f.st, f.dataDir, archive, "1.5.0"); err != nil {
		t.Fatalf("create backup: %v", err)
	}
	inspection, err := backup.Inspect(archive, t.TempDir(), store.CurrentSchemaVersion())
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}

	// 模拟灾难：库与图片都没了。
	if err := f.st.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	for _, target := range []string{f.dbPath, f.dbPath + "-wal", f.dbPath + "-shm", f.mediaPath()} {
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remove %s: %v", target, err)
		}
	}

	if err := backup.Apply(ctx, inspection.ExtractDir, f.dataDir, f.dbPath); err != nil {
		t.Fatalf("apply backup: %v", err)
	}

	next, err := store.Open(ctx, f.dbPath)
	if err != nil {
		t.Fatalf("reopen restored database: %v", err)
	}
	defer next.Close()

	notes, projects, err := next.ContentCounts(ctx)
	if err != nil {
		t.Fatalf("count restored content: %v", err)
	}
	if notes != 1 || projects != 1 {
		t.Fatalf("expected 1 note and 1 project after restore, got %d / %d", notes, projects)
	}
	if _, err := os.Stat(f.mediaPath()); err != nil {
		t.Fatalf("expected media restored: %v", err)
	}
	refs, err := next.MediaReferences(ctx)
	if err != nil {
		t.Fatalf("media references: %v", err)
	}
	if missing := backup.MissingMedia(f.dataDir, refs); len(missing) != 0 {
		t.Fatalf("expected all media present after restore, missing: %v", missing)
	}
}

func writeTarEntry(t *testing.T, tw *tar.Writer, name string, data []byte) {
	t.Helper()
	if err := tw.WriteHeader(&tar.Header{
		Name:     name,
		Mode:     0o600,
		Size:     int64(len(data)),
		Typeflag: tar.TypeReg,
	}); err != nil {
		t.Fatalf("write header %s: %v", name, err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
