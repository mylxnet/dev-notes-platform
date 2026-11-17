package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/model"
	"dev-notes-platform/internal/store"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		_ = st.Close()
	})

	project := &model.Project{
		Name:      "SiteGuard",
		Slug:      "siteguard",
		Summary:   "站点巡检工具",
		TechStack: []string{"Go"},
		RepoURL:   "https://github.com/example/siteguard",
		Status:    model.StatusPublished,
	}
	if _, err := st.CreateProject(context.Background(), project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	published := &model.Note{
		Title:     "公开笔记",
		Slug:      "public-note",
		Summary:   "摘要",
		Content:   "正文",
		ProjectID: &project.ID,
		Status:    model.StatusPublished,
	}
	if _, err := st.CreateNote(context.Background(), published); err != nil {
		t.Fatalf("create published note: %v", err)
	}

	draft := &model.Note{
		Title:   "草稿笔记",
		Slug:    "draft-note",
		Summary: "摘要",
		Content: "正文",
		Status:  model.StatusDraft,
	}
	if _, err := st.CreateNote(context.Background(), draft); err != nil {
		t.Fatalf("create draft note: %v", err)
	}

	return New(st, "test")
}

func TestHealth(t *testing.T) {
	router := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if body["status"] != "ok" || body["db"] != "ok" || body["version"] != "test" {
		t.Fatalf("unexpected health body: %+v", body)
	}
}

func TestPublicNotesOnlyPublished(t *testing.T) {
	router := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/notes", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var body struct {
		Data struct {
			Items    []model.Note `json:"items"`
			Total    int          `json:"total"`
			Page     int          `json:"page"`
			PageSize int          `json:"page_size"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode notes: %v", err)
	}
	if body.Data.Total != 1 || len(body.Data.Items) != 1 || body.Data.Items[0].Slug != "public-note" {
		t.Fatalf("expected only published note, got %+v", body.Data)
	}
	if body.Data.Items[0].Project == nil || body.Data.Items[0].Project.Slug != "siteguard" {
		t.Fatalf("expected project summary, got %+v", body.Data.Items[0].Project)
	}
}

func TestDraftNoteIsNotFound(t *testing.T) {
	router := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/notes/draft-note", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestUnknownAPIReturnsJSON404(t *testing.T) {
	router := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/missing", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %+v", body.Error)
	}
}

func TestProjectDetailIncludesPublishedNotes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/projects/siteguard", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var body struct {
		Data struct {
			Project   model.Project `json:"project"`
			Notes     []model.Note  `json:"notes"`
			NoteCount int           `json:"note_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode project: %v", err)
	}
	if body.Data.Project.Slug != "siteguard" || body.Data.NoteCount != 1 || len(body.Data.Notes) != 1 {
		t.Fatalf("unexpected project detail: %+v", body.Data)
	}
}
