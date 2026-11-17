package store

import (
	"context"
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
