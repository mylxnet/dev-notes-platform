package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"dev-notes-platform/internal/model"
)

// TestSnapshotIncludesWALData 验证快照包含 WAL 中尚未 checkpoint 的写入。
// 这是「直接复制 notes.db」会丢数据的原因，也是备份必须走 VACUUM INTO 的依据。
func TestSnapshotIncludesWALData(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "notes.db")

	st, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	// 关掉自动 checkpoint，强制数据留在 -wal 文件里。
	if _, err := st.db.ExecContext(ctx, "PRAGMA wal_autocheckpoint=0"); err != nil {
		t.Fatalf("disable autocheckpoint: %v", err)
	}

	note := &model.Note{
		Title:   "只写在 WAL 里的笔记",
		Slug:    "wal-only",
		Summary: "摘要",
		Content: "正文",
		Status:  model.StatusPublished,
	}
	if _, err := st.CreateNote(ctx, note); err != nil {
		t.Fatalf("create note: %v", err)
	}
	if _, err := os.Stat(dbPath + "-wal"); err != nil {
		t.Fatalf("expected wal file to exist: %v", err)
	}

	snapPath := filepath.Join(dir, "snapshot.db")
	if err := st.SnapshotTo(ctx, snapPath); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	raw, err := sql.Open("sqlite", "file:"+snapPath+"?mode=ro")
	if err != nil {
		t.Fatalf("open snapshot: %v", err)
	}
	defer raw.Close()

	var count int
	if err := raw.QueryRow(`SELECT COUNT(*) FROM note WHERE slug = 'wal-only'`).Scan(&count); err != nil {
		t.Fatalf("count snapshot rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected note to be present in snapshot, got %d", count)
	}

	var integrity string
	if err := raw.QueryRow("PRAGMA integrity_check").Scan(&integrity); err != nil {
		t.Fatalf("integrity check: %v", err)
	}
	if integrity != "ok" {
		t.Fatalf("expected ok integrity, got %q", integrity)
	}
}

// TestMediaReferences 验证从内容里能收集到上传图片引用。
func TestMediaReferences(t *testing.T) {
	ctx := context.Background()
	st := newTestStore(t)

	project := &model.Project{
		Name:        "带封面",
		Slug:        "with-cover",
		Summary:     "摘要",
		RepoURL:     "https://github.com/example/cover",
		CoverImage:  "/media/uploads/" + testHash("a") + ".png",
		Description: "![](/media/uploads/" + testHash("b") + ".png)",
		Status:      model.StatusPublished,
	}
	if _, err := st.CreateProject(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}
	note := &model.Note{
		Title:   "带插图",
		Slug:    "with-image",
		Summary: "摘要",
		Content: "![](/media/uploads/" + testHash("a") + ".png) 和 /media/media-placeholder.svg",
		Status:  model.StatusPublished,
	}
	if _, err := st.CreateNote(ctx, note); err != nil {
		t.Fatalf("create note: %v", err)
	}

	refs, err := st.MediaReferences(ctx)
	if err != nil {
		t.Fatalf("media references: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 unique upload references, got %v", refs)
	}
}

// testHash 生成 32 位占位哈希，避免测试里写死长串。
func testHash(seed string) string {
	out := ""
	for len(out) < 32 {
		out += seed
	}
	return out[:32]
}
