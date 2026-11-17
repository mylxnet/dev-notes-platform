package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/auth"
	"dev-notes-platform/internal/store"
)

const (
	testAdminUser     = "admin"
	testAdminPassword = "admin12345678"
)

func newAdminTestServer(t *testing.T) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "admin.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if err := st.SeedAdmin(context.Background(), testAdminUser, testAdminPassword); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	return New(st, "test", auth.NewManager(strings.Repeat("t", 32), auth.CookieSecureNever))
}

func doJSON(t *testing.T, router http.Handler, method, path, payload string, cookies []*http.Cookie, csrf string) *httptest.ResponseRecorder {
	t.Helper()
	var body *strings.Reader
	if payload == "" {
		body = strings.NewReader("")
	} else {
		body = strings.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	if csrf != "" {
		req.Header.Set(auth.CSRFHeader, csrf)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// login 返回会话 cookie 与 CSRF token。
func login(t *testing.T, router http.Handler) ([]*http.Cookie, string) {
	t.Helper()
	payload := `{"username":"` + testAdminUser + `","password":"` + testAdminPassword + `"}`
	rec := doJSON(t, router, http.MethodPost, "/api/admin/login", payload, nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			CSRFToken string `json:"csrf_token"`
			Admin     struct {
				Username string `json:"username"`
			} `json:"admin"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if body.Data.CSRFToken == "" || body.Data.Admin.Username != testAdminUser {
		t.Fatalf("unexpected login body: %s", rec.Body.String())
	}
	return rec.Result().Cookies(), body.Data.CSRFToken
}

func TestAdminRequiresLogin(t *testing.T) {
	router := newAdminTestServer(t)
	rec := doJSON(t, router, http.MethodGet, "/api/admin/notes", "", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAdminLoginRejectsWrongPassword(t *testing.T) {
	router := newAdminTestServer(t)
	rec := doJSON(t, router, http.MethodPost, "/api/admin/login",
		`{"username":"admin","password":"wrong-password-1"}`, nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "账号或密码不正确") {
		t.Fatalf("expected unified message, got %s", rec.Body.String())
	}
}

func TestAdminWriteRequiresCSRF(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, _ := login(t, router)

	payload := `{"title":"标题","slug":"csrf-note","summary":"摘要","content":"正文","status":"draft"}`
	rec := doJSON(t, router, http.MethodPost, "/api/admin/notes", payload, cookies, "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without csrf, got %d", rec.Code)
	}
}

func TestAdminCreateNoteFlow(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	payload := `{"title":"新笔记","slug":"new-note","summary":"摘要","content":"# 正文\n\n段落","status":"published","published_at":"2026-11-17T00:00:00Z"}`
	rec := doJSON(t, router, http.MethodPost, "/api/admin/notes", payload, cookies, csrf)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "new-note") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}

	// 前台可见
	public := doJSON(t, router, http.MethodGet, "/api/notes/new-note", "", nil, "")
	if public.Code != http.StatusOK {
		t.Fatalf("expected public note, got %d", public.Code)
	}
	if !strings.Contains(public.Body.String(), "content_html") {
		t.Fatalf("expected content_html, got %s", public.Body.String())
	}

	// 重复 slug
	dup := doJSON(t, router, http.MethodPost, "/api/admin/notes", payload, cookies, csrf)
	if dup.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", dup.Code)
	}
}

func TestAdminDraftNotPublic(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	payload := `{"title":"草稿","slug":"draft-only","summary":"摘要","content":"正文","status":"draft"}`
	rec := doJSON(t, router, http.MethodPost, "/api/admin/notes", payload, cookies, csrf)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d %s", rec.Code, rec.Body.String())
	}

	public := doJSON(t, router, http.MethodGet, "/api/notes/draft-only", "", nil, "")
	if public.Code != http.StatusNotFound {
		t.Fatalf("expected draft hidden from public, got %d", public.Code)
	}
}

func TestAdminBlocksSensitiveWording(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	payload := `{"title":"说明","slug":"blocked-note","summary":"摘要","content":"完整版请从网盘地址获取安装包","status":"draft"}`
	rec := doJSON(t, router, http.MethodPost, "/api/admin/notes", payload, cookies, csrf)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code       string   `json:"code"`
			Field      string   `json:"field"`
			Hits       []string `json:"hits"`
			Suggestion string   `json:"suggestion"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "CONTENT_BLOCKED" || body.Error.Field != "content" ||
		len(body.Error.Hits) == 0 || body.Error.Suggestion == "" {
		t.Fatalf("unexpected blocked body: %s", rec.Body.String())
	}
}

func TestAdminProjectCRUD(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	payload := `{"name":"示例项目","slug":"demo-project","summary":"摘要","tech_stack":["Go"],"repo_url":"https://github.com/example/demo","sort_order":1,"status":"published"}`
	created := doJSON(t, router, http.MethodPost, "/api/admin/projects", payload, cookies, csrf)
	if created.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d %s", created.Code, created.Body.String())
	}

	var body struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	updatePayload := `{"name":"示例项目改名","slug":"demo-project","summary":"新摘要","tech_stack":["Go","Vue"],"repo_url":"https://github.com/example/demo","sort_order":2,"status":"draft"}`
	updated := doJSON(t, router, http.MethodPut, "/api/admin/projects/"+itoa(body.Data.ID), updatePayload, cookies, csrf)
	if updated.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", updated.Code, updated.Body.String())
	}

	list := doJSON(t, router, http.MethodGet, "/api/admin/projects", "", cookies, "")
	if !strings.Contains(list.Body.String(), "示例项目改名") {
		t.Fatalf("expected updated name, got %s", list.Body.String())
	}

	deleted := doJSON(t, router, http.MethodDelete, "/api/admin/projects/"+itoa(body.Data.ID), "", cookies, csrf)
	if deleted.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", deleted.Code)
	}
}

func TestAdminPasswordChangeInvalidatesOldSession(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	payload := `{"current_password":"` + testAdminPassword + `","new_password":"newadmin123456"}`
	rec := doJSON(t, router, http.MethodPut, "/api/admin/password", payload, cookies, csrf)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", rec.Code, rec.Body.String())
	}

	// 旧 cookie 应失效
	old := doJSON(t, router, http.MethodGet, "/api/admin/me", "", cookies, "")
	if old.Code != http.StatusUnauthorized {
		t.Fatalf("expected old session invalid, got %d", old.Code)
	}
}

func TestAdminLoginRateLimited(t *testing.T) {
	router := newAdminTestServer(t)
	for i := 0; i < 5; i++ {
		rec := doJSON(t, router, http.MethodPost, "/api/admin/login",
			`{"username":"admin","password":"wrong-password-1"}`, nil, "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, rec.Code)
		}
	}
	rec := doJSON(t, router, http.MethodPost, "/api/admin/login",
		`{"username":"admin","password":"wrong-password-1"}`, nil, "")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after limit, got %d", rec.Code)
	}
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
