package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
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
	handler, _ := newAdminTestServerWithMedia(t)
	return handler
}

func newAdminTestServerWithMedia(t *testing.T) (http.Handler, string) {
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

	mediaDir := filepath.Join(t.TempDir(), "media")
	return New(st, "test", auth.NewManager(strings.Repeat("t", 32), auth.CookieSecureNever), mediaDir), mediaDir
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

	payload := `{"name":"示例项目","slug":"demo-project","summary":"摘要","tech_stack":["Go"],"repo_url":"https://github.com/example/demo","description":"## 定位\n\n练习项目。","cover_image":"/media/projects/demo.png","sort_order":1,"status":"published"}`
	created := doJSON(t, router, http.MethodPost, "/api/admin/projects", payload, cookies, csrf)
	if created.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d %s", created.Code, created.Body.String())
	}

	var body struct {
		Data struct {
			ID          int64  `json:"id"`
			Description string `json:"description"`
			CoverImage  string `json:"cover_image"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Data.Description != "## 定位\n\n练习项目。" || body.Data.CoverImage != "/media/projects/demo.png" {
		t.Fatalf("expected description and cover persisted, got %+v", body.Data)
	}

	updatePayload := `{"name":"示例项目改名","slug":"demo-project","summary":"新摘要","tech_stack":["Go","Vue"],"repo_url":"https://github.com/example/demo","description":"更新后的说明","cover_image":"","sort_order":2,"status":"draft"}`
	updated := doJSON(t, router, http.MethodPut, "/api/admin/projects/"+itoa(body.Data.ID), updatePayload, cookies, csrf)
	if updated.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", updated.Code, updated.Body.String())
	}
	if !strings.Contains(updated.Body.String(), "更新后的说明") {
		t.Fatalf("expected updated description, got %s", updated.Body.String())
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

// TestAdminProjectListPaging 验证后台项目列表按 page/page_size 分页，并返回总数供前端算页码。
func TestAdminProjectListPaging(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	for index := 1; index <= 7; index++ {
		slug := "paging-" + string(rune('a'+index))
		payload := `{"name":"分页项目` + itoa(int64(index)) + `","slug":"` + slug +
			`","summary":"摘要","tech_stack":["Go"],"repo_url":"","description":"","cover_image":"","sort_order":` +
			itoa(int64(index)) + `,"status":"draft"}`
		rec := doJSON(t, router, http.MethodPost, "/api/admin/projects", payload, cookies, csrf)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create project %d: expected 201, got %d %s", index, rec.Code, rec.Body.String())
		}
	}

	var body struct {
		Data struct {
			Items []struct {
				SortOrder int `json:"sort_order"`
			} `json:"items"`
			Total    int `json:"total"`
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
		} `json:"data"`
	}

	first := doJSON(t, router, http.MethodGet, "/api/admin/projects?page=1&page_size=6", "", cookies, "")
	if first.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", first.Code, first.Body.String())
	}
	if err := json.Unmarshal(first.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode first page: %v", err)
	}
	if body.Data.Total != 7 || body.Data.Page != 1 || body.Data.PageSize != 6 || len(body.Data.Items) != 6 {
		t.Fatalf("unexpected first page: %s", first.Body.String())
	}
	if body.Data.Items[0].SortOrder != 1 || body.Data.Items[5].SortOrder != 6 {
		t.Fatalf("expected sort_order 1..6 on first page, got %s", first.Body.String())
	}

	second := doJSON(t, router, http.MethodGet, "/api/admin/projects?page=2&page_size=6", "", cookies, "")
	if err := json.Unmarshal(second.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode second page: %v", err)
	}
	if body.Data.Total != 7 || body.Data.Page != 2 || len(body.Data.Items) != 1 || body.Data.Items[0].SortOrder != 7 {
		t.Fatalf("unexpected second page: %s", second.Body.String())
	}

	// 不传参数时按默认 20 条返回，老前端调用方式不受影响。
	all := doJSON(t, router, http.MethodGet, "/api/admin/projects", "", cookies, "")
	if err := json.Unmarshal(all.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode default page: %v", err)
	}
	if len(body.Data.Items) != 7 || body.Data.PageSize != 20 {
		t.Fatalf("unexpected default page: %s", all.Body.String())
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

// TestMediaUpload 验证封面图上传的类型/大小校验、随机文件名与运行时下发。
func TestMediaUpload(t *testing.T) {
	router, mediaDir := newAdminTestServerWithMedia(t)
	cookies, csrf := login(t, router)

	// 1×1 透明 PNG
	pngBytes, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFAAH/q842iQAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatalf("decode png fixture: %v", err)
	}

	upload := func(filename string, data []byte, useCookies bool, token string) *httptest.ResponseRecorder {
		t.Helper()
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(data); err != nil {
			t.Fatalf("write form file: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close multipart: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/admin/media/upload", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		if useCookies {
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
		}
		if token != "" {
			req.Header.Set(auth.CSRFHeader, token)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	unauth := upload("cover.png", pngBytes, false, "")
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without login, got %d", unauth.Code)
	}
	noCSRF := upload("cover.png", pngBytes, true, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without csrf, got %d", noCSRF.Code)
	}
	badExt := upload("cover.svg", pngBytes, true, csrf)
	if badExt.Code != http.StatusUnprocessableEntity || !strings.Contains(badExt.Body.String(), "UNSUPPORTED_IMAGE") {
		t.Fatalf("expected 422 for svg, got %d %s", badExt.Code, badExt.Body.String())
	}
	fake := upload("cover.png", []byte("not an image, just text"), true, csrf)
	if fake.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for fake image, got %d %s", fake.Code, fake.Body.String())
	}
	tooBig := upload("cover.png", bytes.Repeat([]byte("x"), maxUploadBytes+1), true, csrf)
	if tooBig.Code != http.StatusUnprocessableEntity || !strings.Contains(tooBig.Body.String(), "IMAGE_TOO_LARGE") {
		t.Fatalf("expected 422 for oversized file, got %d %s", tooBig.Code, tooBig.Body.String())
	}

	ok := upload("cover.png", pngBytes, true, csrf)
	if ok.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", ok.Code, ok.Body.String())
	}
	var body struct {
		Data struct {
			Path string `json:"path"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ok.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	imagePath := body.Data.Path
	if !regexp.MustCompile(`^/media/uploads/[a-f0-9]{32}\.png$`).MatchString(imagePath) {
		t.Fatalf("unexpected image path: %q", imagePath)
	}
	if _, err := os.Stat(filepath.Join(mediaDir, strings.TrimPrefix(imagePath, "/media/uploads/"))); err != nil {
		t.Fatalf("uploaded file missing: %v", err)
	}

	served := doJSON(t, router, http.MethodGet, imagePath, "", nil, "")
	if served.Code != http.StatusOK || !strings.HasPrefix(served.Header().Get("Content-Type"), "image/png") {
		t.Fatalf("expected served png, got %d %q", served.Code, served.Header().Get("Content-Type"))
	}
	traversal := doJSON(t, router, http.MethodGet, "/media/uploads/not-a-valid-name.png", "", nil, "")
	if traversal.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for bad name, got %d", traversal.Code)
	}
}

// TestSiteSettings 验证备案号公开读取与后台保存：非法值拦截、清空恢复占位。
func TestSiteSettings(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	initial := doJSON(t, router, http.MethodGet, "/api/site", "", nil, "")
	if initial.Code != http.StatusOK || !strings.Contains(initial.Body.String(), `"icp_number":""`) {
		t.Fatalf("expected empty icp_number, got %d %s", initial.Code, initial.Body.String())
	}

	unauth := doJSON(t, router, http.MethodGet, "/api/admin/settings", "", nil, "")
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without login, got %d", unauth.Code)
	}
	noCSRF := doJSON(t, router, http.MethodPut, "/api/admin/settings", `{"icp_number":"甘ICP备12345678号"}`, cookies, "")
	if noCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without csrf, got %d", noCSRF.Code)
	}

	bad := doJSON(t, router, http.MethodPut, "/api/admin/settings", `{"icp_number":"<script>x</script>"}`, cookies, csrf)
	if bad.Code != http.StatusUnprocessableEntity || !strings.Contains(bad.Body.String(), "icp_number") {
		t.Fatalf("expected 422 with field, got %d %s", bad.Code, bad.Body.String())
	}

	saved := doJSON(t, router, http.MethodPut, "/api/admin/settings", `{"icp_number":" 甘ICP备12345678号-1 "}`, cookies, csrf)
	if saved.Code != http.StatusOK || !strings.Contains(saved.Body.String(), "甘ICP备12345678号-1") {
		t.Fatalf("expected saved icp_number, got %d %s", saved.Code, saved.Body.String())
	}
	adminGet := doJSON(t, router, http.MethodGet, "/api/admin/settings", "", cookies, "")
	if !strings.Contains(adminGet.Body.String(), "甘ICP备12345678号-1") {
		t.Fatalf("expected admin get to read value, got %s", adminGet.Body.String())
	}
	public := doJSON(t, router, http.MethodGet, "/api/site", "", nil, "")
	if !strings.Contains(public.Body.String(), "甘ICP备12345678号-1") {
		t.Fatalf("expected public site to read value, got %s", public.Body.String())
	}

	cleared := doJSON(t, router, http.MethodPut, "/api/admin/settings", `{"icp_number":""}`, cookies, csrf)
	if cleared.Code != http.StatusOK {
		t.Fatalf("expected 200 on clear, got %d %s", cleared.Code, cleared.Body.String())
	}
	public = doJSON(t, router, http.MethodGet, "/api/site", "", nil, "")
	if !strings.Contains(public.Body.String(), `"icp_number":""`) {
		t.Fatalf("expected empty after clear, got %s", public.Body.String())
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
