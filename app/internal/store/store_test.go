package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"dev-notes-platform/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		_ = st.Close()
	})
	return st
}

func TestPublishedContentFiltering(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	project := &model.Project{
		Name:      "SuperClip-C",
		Slug:      "superclip",
		Summary:   "剪贴板历史管理工具",
		TechStack: []string{"Go", "WPF"},
		RepoURL:   "https://github.com/example/superclip",
		Status:    model.StatusPublished,
	}
	if _, err := st.CreateProject(ctx, project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	published := &model.Note{
		Title:     "已发布笔记",
		Slug:      "published-note",
		Summary:   "摘要",
		Content:   "正文",
		ProjectID: &project.ID,
		Status:    model.StatusPublished,
	}
	if _, err := st.CreateNote(ctx, published); err != nil {
		t.Fatalf("create published note: %v", err)
	}

	draft := &model.Note{
		Title:   "草稿笔记",
		Slug:    "draft-note",
		Summary: "摘要",
		Content: "正文",
		Status:  model.StatusDraft,
	}
	if _, err := st.CreateNote(ctx, draft); err != nil {
		t.Fatalf("create draft note: %v", err)
	}

	notes, err := st.ListPublishedNotes(ctx, NoteFilter{})
	if err != nil {
		t.Fatalf("list published notes: %v", err)
	}
	if len(notes) != 1 || notes[0].Slug != "published-note" {
		t.Fatalf("expected only published note, got %+v", notes)
	}
	if notes[0].Project == nil || notes[0].Project.Slug != "superclip" {
		t.Fatalf("expected project summary, got %+v", notes[0].Project)
	}

	if _, err := st.GetPublishedNoteBySlug(ctx, "draft-note"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("draft should not be public, got %v", err)
	}

	count, err := st.CountPublishedNotesByProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("count project notes: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 published project note, got %d", count)
	}

	projects, err := st.ListPublishedProjects(ctx)
	if err != nil {
		t.Fatalf("list published projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected one published project, got %+v", projects)
	}
	if projects[0].NoteCount != 1 {
		t.Fatalf("expected project note_count=1, got %d", projects[0].NoteCount)
	}
}

func TestAdminSeedAndPasswordHash(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	if err := st.SeedAdmin(ctx, "admin", "admin12345678"); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	if err := st.SeedAdmin(ctx, "admin", "another-password"); err != nil {
		t.Fatalf("seed admin again: %v", err)
	}

	count, err := st.AdminCount(ctx)
	if err != nil {
		t.Fatalf("admin count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 admin, got %d", count)
	}

	admin, err := st.GetAdminByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("get admin: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte("admin12345678")); err != nil {
		t.Fatalf("password hash mismatch: %v", err)
	}
}

func TestSlugConflict(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	first := &model.Project{Name: "A", Slug: "same-slug", Summary: "A", RepoURL: "https://github.com/a/a", Status: model.StatusDraft}
	if _, err := st.CreateProject(ctx, first); err != nil {
		t.Fatalf("create first project: %v", err)
	}
	second := &model.Project{Name: "B", Slug: "same-slug", Summary: "B", RepoURL: "https://github.com/b/b", Status: model.StatusDraft}
	if _, err := st.CreateProject(ctx, second); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

// TestMigrateV1Database 验证 v1 老库升级到 v2 时补上说明与封面列，且既有数据不丢。
func TestMigrateV1Database(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v1.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open raw sqlite: %v", err)
	}
	// 按 v1 的建表语句造一份没有 description / cover_image 的老库。
	for _, statement := range []string{
		`CREATE TABLE project (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			summary TEXT NOT NULL,
			tech_stack TEXT NOT NULL DEFAULT '[]',
			repo_url TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'draft',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE note (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			summary TEXT NOT NULL,
			content TEXT NOT NULL,
			project_id INTEGER,
			status TEXT NOT NULL DEFAULT 'draft',
			published_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`INSERT INTO project (name, slug, summary, tech_stack, repo_url, created_at, updated_at)
			VALUES ('老项目', 'old-project', '摘要', '["Go"]', 'https://github.com/a/b', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');`,
		`PRAGMA user_version = 1;`,
	} {
		if _, err := raw.Exec(statement); err != nil {
			_ = raw.Close()
			t.Fatalf("prepare v1 schema: %v", err)
		}
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close raw sqlite: %v", err)
	}

	st, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("migrate v1 database: %v", err)
	}
	defer st.Close()
	ctx := context.Background()

	project, err := st.GetProjectByID(ctx, 1)
	if err != nil {
		t.Fatalf("read migrated project: %v", err)
	}
	if project.Name != "老项目" {
		t.Fatalf("expected v1 row kept, got %+v", project)
	}
	if project.Description != "" || project.CoverImage != "" {
		t.Fatalf("expected empty new columns, got %q / %q", project.Description, project.CoverImage)
	}

	// 新列可写可读。
	project.Description = "老库补列后的说明"
	project.CoverImage = "/media/projects/old-project.png"
	if err := st.UpdateProject(ctx, project); err != nil {
		t.Fatalf("update migrated project: %v", err)
	}
	again, err := st.GetProjectByID(ctx, 1)
	if err != nil {
		t.Fatalf("reread project: %v", err)
	}
	if again.Description != project.Description || again.CoverImage != project.CoverImage {
		t.Fatalf("expected new columns persisted, got %+v", again)
	}

	// 再开一次不应重复执行 ALTER 而报错。
	second, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen migrated database: %v", err)
	}
	// v1 老库一路升到 v3 后，settings 表应可直接读取。
	if _, err := second.GetSetting(ctx, "icp_number"); err != nil {
		t.Fatalf("settings table missing after migration: %v", err)
	}
	_ = second.Close()
}

// TestMigrateFeaturedBackfill 验证 v3 → v4：旧版存在站点设置里的推荐位会回填成 note.is_featured，
// 且重复打开不再重复执行迁移。
func TestMigrateFeaturedBackfill(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open raw sqlite: %v", err)
	}
	for _, statement := range []string{
		`CREATE TABLE project (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			summary TEXT NOT NULL,
			tech_stack TEXT NOT NULL DEFAULT '[]',
			repo_url TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			cover_image TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'draft',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE note (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			summary TEXT NOT NULL,
			content TEXT NOT NULL,
			project_id INTEGER,
			status TEXT NOT NULL DEFAULT 'draft',
			published_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL
		);`,
		`INSERT INTO note (title, slug, summary, content, status, published_at, created_at, updated_at)
			VALUES ('老推荐', 'old-featured', '摘要', '正文', 'published', '2026-11-01T00:00:00Z', '2026-11-01T00:00:00Z', '2026-11-01T00:00:00Z');`,
		`INSERT INTO note (title, slug, summary, content, status, published_at, created_at, updated_at)
			VALUES ('普通笔记', 'old-plain', '摘要', '正文', 'published', '2026-11-02T00:00:00Z', '2026-11-02T00:00:00Z', '2026-11-02T00:00:00Z');`,
		`INSERT INTO settings (key, value, updated_at)
			VALUES ('featured_notes', '["old-featured"]', '2026-11-01T00:00:00Z');`,
		`PRAGMA user_version = 3;`,
	} {
		if _, err := raw.Exec(statement); err != nil {
			_ = raw.Close()
			t.Fatalf("prepare v3 schema: %v", err)
		}
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close raw sqlite: %v", err)
	}

	st, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("migrate v3 database: %v", err)
	}
	ctx := context.Background()

	notes, err := st.ListAllNotes(ctx, AdminNoteFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list notes after migration: %v", err)
	}
	featured := map[string]bool{}
	for _, note := range notes {
		featured[note.Slug] = note.IsFeatured
	}
	if !featured["old-featured"] {
		t.Fatalf("expected old-featured backfilled as featured, got %+v", featured)
	}
	if featured["old-plain"] {
		t.Fatalf("expected old-plain to stay unfeatured, got %+v", featured)
	}

	// 迁移后新列可写可读。
	note, err := st.GetNoteByID(ctx, notes[0].ID)
	if err != nil {
		t.Fatalf("read note after migration: %v", err)
	}
	if note.Slug == "old-plain" {
		note.IsFeatured = true
		if err := st.UpdateNote(ctx, note); err != nil {
			t.Fatalf("update featured flag: %v", err)
		}
		reread, err := st.GetNoteByID(ctx, note.ID)
		if err != nil {
			t.Fatalf("reread note: %v", err)
		}
		if !reread.IsFeatured {
			t.Fatalf("expected featured persisted, got %+v", reread)
		}
	}

	// 再开一次不应重复执行 ALTER 或重复回填。
	second, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen migrated database: %v", err)
	}
	if _, err := second.ListFeaturedNotes(ctx, 10); err != nil {
		t.Fatalf("list featured after reopen: %v", err)
	}
	_ = second.Close()
}

// TestSettingsStore 验证站点配置读写：未配置为普通空串，同键覆盖更新。
func TestSettingsStore(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	value, err := st.GetSetting(ctx, "icp_number")
	if err != nil {
		t.Fatalf("get unset setting: %v", err)
	}
	if value != "" {
		t.Fatalf("expected empty default, got %q", value)
	}

	for _, next := range []string{"甘ICP备12345678号-1", "甘ICP备87654321号"} {
		if err := st.SetSetting(ctx, "icp_number", next); err != nil {
			t.Fatalf("set setting: %v", err)
		}
		value, err = st.GetSetting(ctx, "icp_number")
		if err != nil {
			t.Fatalf("get setting: %v", err)
		}
		if value != next {
			t.Fatalf("expected %q, got %q", next, value)
		}
	}

	if err := st.SetSetting(ctx, "icp_number", ""); err != nil {
		t.Fatalf("clear setting: %v", err)
	}
	value, err = st.GetSetting(ctx, "icp_number")
	if err != nil || value != "" {
		t.Fatalf("expected cleared setting, got %q err=%v", value, err)
	}
}

// TestAdminProjectPaging 验证后台项目列表按「排序值 + ID」稳定分页，且总数统计独立于分页窗口。
func TestAdminProjectPaging(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	for index := 1; index <= 7; index++ {
		project := &model.Project{
			Name:      "项目",
			Slug:      "paging-project-" + string(rune('a'+index)),
			Summary:   "摘要",
			RepoURL:   "https://github.com/example/paging",
			SortOrder: index,
			Status:    model.StatusDraft,
		}
		if _, err := st.CreateProject(ctx, project); err != nil {
			t.Fatalf("create project %d: %v", index, err)
		}
	}

	total, err := st.CountAllProjects(ctx)
	if err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if total != 7 {
		t.Fatalf("expected total 7, got %d", total)
	}

	first, err := st.ListAllProjects(ctx, AdminProjectFilter{Limit: 6, Offset: 0})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(first) != 6 {
		t.Fatalf("expected 6 projects on first page, got %d", len(first))
	}
	if first[0].SortOrder != 1 || first[5].SortOrder != 6 {
		t.Fatalf("expected sort_order 1..6, got %d..%d", first[0].SortOrder, first[5].SortOrder)
	}

	last, err := st.ListAllProjects(ctx, AdminProjectFilter{Limit: 6, Offset: 6})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(last) != 1 || last[0].SortOrder != 7 {
		t.Fatalf("expected one project with sort_order 7 on second page, got %+v", last)
	}

	// 越界页返回空列表而不是报错，前端据此退回最后一页。
	beyond, err := st.ListAllProjects(ctx, AdminProjectFilter{Limit: 6, Offset: 12})
	if err != nil {
		t.Fatalf("list beyond last page: %v", err)
	}
	if len(beyond) != 0 {
		t.Fatalf("expected empty page beyond last, got %d items", len(beyond))
	}
}
