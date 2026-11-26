package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dev-notes-platform/internal/auth"
)

// doMultipart 发送带文件的 multipart 请求，模拟后台上传备份包。
func doMultipart(t *testing.T, router http.Handler, path, filename string, content []byte,
	extra map[string]string, cookies []*http.Cookie, csrf string) *httptest.ResponseRecorder {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	for key, value := range extra {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field %s: %v", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
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

// downloadBackup 调下载接口，把响应体落盘后返回路径。
func downloadBackup(t *testing.T, router http.Handler, cookies []*http.Cookie, csrf, dir string) string {
	t.Helper()
	rec := doJSON(t, router, http.MethodPost, "/api/admin/backup", "", cookies, csrf)
	if rec.Code != http.StatusOK {
		t.Fatalf("backup download failed: %d %s", rec.Code, rec.Body.String())
	}
	if disposition := rec.Header().Get("Content-Disposition"); !strings.Contains(disposition, ".tar.gz") {
		t.Fatalf("unexpected content disposition: %q", disposition)
	}
	path := filepath.Join(dir, "backup.tar.gz")
	if err := os.WriteFile(path, rec.Body.Bytes(), 0o600); err != nil {
		t.Fatalf("write backup file: %v", err)
	}
	return path
}

// TestBackupRequiresLogin 验证备份接口需要登录。
func TestBackupRequiresLogin(t *testing.T) {
	router := newAdminTestServer(t)
	rec := doJSON(t, router, http.MethodGet, "/api/admin/backup/summary", "", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// TestBackupDownloadRequiresCSRF 验证 POST 下载缺少 CSRF 时被拒绝。
func TestBackupDownloadRequiresCSRF(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, _ := login(t, router)
	rec := doJSON(t, router, http.MethodPost, "/api/admin/backup", "", cookies, "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d %s", rec.Code, rec.Body.String())
	}
}

// TestBackupSummaryAndDownload 验证概览数据与下载的包能被预检通过。
func TestBackupSummaryAndDownload(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	summary := doJSON(t, router, http.MethodGet, "/api/admin/backup/summary", "", cookies, "")
	if summary.Code != http.StatusOK {
		t.Fatalf("summary failed: %d %s", summary.Code, summary.Body.String())
	}
	if !strings.Contains(summary.Body.String(), `"schema_version":4`) {
		t.Fatalf("expected schema version in summary, got %s", summary.Body.String())
	}

	archivePath := downloadBackup(t, router, cookies, csrf, t.TempDir())
	content, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}

	rec := doMultipart(t, router, "/api/admin/backup/inspect", "backup.tar.gz", content, nil, cookies, csrf)
	if rec.Code != http.StatusOK {
		t.Fatalf("inspect failed: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			Restorable bool     `json:"restorable"`
			Problems   []string `json:"problems"`
			Missing    []string `json:"missing"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode inspect: %v", err)
	}
	if !body.Data.Restorable || len(body.Data.Problems) != 0 {
		t.Fatalf("expected restorable backup, got %s", rec.Body.String())
	}
	if len(body.Data.Missing) != 0 {
		t.Fatalf("expected no missing media, got %v", body.Data.Missing)
	}
}

// TestRestoreFlow 验证「恢复覆盖当前数据」，且恢复前快照可以回滚。
func TestRestoreFlow(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	payload := `{"title":"备份测试","slug":"backup-test","summary":"摘要","content":"正文",` +
		`"status":"published","published_at":"2026-11-27T00:00:00Z","is_featured":false}`
	created := doJSON(t, router, http.MethodPost, "/api/admin/notes", payload, cookies, csrf)
	if created.Code != http.StatusCreated {
		t.Fatalf("create note failed: %d %s", created.Code, created.Body.String())
	}
	var createdBody struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createdBody); err != nil {
		t.Fatalf("decode created note: %v", err)
	}

	archivePath := downloadBackup(t, router, cookies, csrf, t.TempDir())
	content, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}

	// 删掉笔记，模拟数据丢失。
	deleted := doJSON(t, router, http.MethodDelete,
		"/api/admin/notes/"+itoa(createdBody.Data.ID), "", cookies, csrf)
	if deleted.Code != http.StatusNoContent && deleted.Code != http.StatusOK {
		t.Fatalf("delete note failed: %d %s", deleted.Code, deleted.Body.String())
	}
	before := doJSON(t, router, http.MethodGet, "/api/notes?q=备份测试", "", nil, "")
	if !strings.Contains(before.Body.String(), `"total":0`) {
		t.Fatalf("expected note gone before restore, got %s", before.Body.String())
	}

	// 缺 confirm 时拒绝。
	noConfirm := doMultipart(t, router, "/api/admin/backup/restore", "backup.tar.gz", content, nil, cookies, csrf)
	if noConfirm.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without confirm, got %d %s", noConfirm.Code, noConfirm.Body.String())
	}

	restored := doMultipart(t, router, "/api/admin/backup/restore", "backup.tar.gz", content,
		map[string]string{"confirm": "RESTORE"}, cookies, csrf)
	if restored.Code != http.StatusOK {
		t.Fatalf("restore failed: %d %s", restored.Code, restored.Body.String())
	}
	var restoredBody struct {
		Data struct {
			Notes    int      `json:"notes"`
			Projects int      `json:"projects"`
			Snapshot string   `json:"snapshot"`
			Missing  []string `json:"missing"`
		} `json:"data"`
	}
	if err := json.Unmarshal(restored.Body.Bytes(), &restoredBody); err != nil {
		t.Fatalf("decode restore: %v", err)
	}
	if restoredBody.Data.Notes != 1 || restoredBody.Data.Snapshot == "" {
		t.Fatalf("unexpected restore result: %s", restored.Body.String())
	}
	if len(restoredBody.Data.Missing) != 0 {
		t.Fatalf("expected no missing media after restore, got %v", restoredBody.Data.Missing)
	}

	after := doJSON(t, router, http.MethodGet, "/api/notes?q=备份测试", "", nil, "")
	if !strings.Contains(after.Body.String(), `"total":1`) {
		t.Fatalf("expected note restored, got %s", after.Body.String())
	}

	// 快照回滚：先删掉数据再回滚到恢复前状态（恢复前那一步库里已没有笔记）。
	snapshots := doJSON(t, router, http.MethodGet, "/api/admin/backup/snapshots", "", cookies, "")
	if snapshots.Code != http.StatusOK || !strings.Contains(snapshots.Body.String(), restoredBody.Data.Snapshot) {
		t.Fatalf("expected snapshot listed, got %d %s", snapshots.Code, snapshots.Body.String())
	}

	rollback := doJSON(t, router, http.MethodPost,
		"/api/admin/backup/snapshots/"+restoredBody.Data.Snapshot+"/restore", "", cookies, csrf)
	if rollback.Code != http.StatusOK {
		t.Fatalf("rollback failed: %d %s", rollback.Code, rollback.Body.String())
	}
	// 快照拍在删除笔记之后、恢复之前，因此回滚后库应回到「没有这条笔记」的状态。
	rolledBack := doJSON(t, router, http.MethodGet, "/api/notes?q=备份测试", "", nil, "")
	if !strings.Contains(rolledBack.Body.String(), `"total":0`) {
		t.Fatalf("expected note gone after rollback, got %s", rolledBack.Body.String())
	}
}

// TestRestoreRejectsBrokenArchive 验证乱传的文件被拦住。
func TestRestoreRejectsBrokenArchive(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	rec := doMultipart(t, router, "/api/admin/backup/restore", "bad.tar.gz",
		[]byte("not a gzip archive"), map[string]string{"confirm": "RESTORE"}, cookies, csrf)
	if rec.Code != http.StatusUnprocessableEntity && rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 4xx for broken archive, got %d %s", rec.Code, rec.Body.String())
	}
}

// TestRestoreRejectsWrongExtension 验证只接受 .tar.gz。
func TestRestoreRejectsWrongExtension(t *testing.T) {
	router := newAdminTestServer(t)
	cookies, csrf := login(t, router)

	rec := doMultipart(t, router, "/api/admin/backup/inspect", "backup.zip",
		[]byte("x"), nil, cookies, csrf)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for wrong extension, got %d %s", rec.Code, rec.Body.String())
	}
}
