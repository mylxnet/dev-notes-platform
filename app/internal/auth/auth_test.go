package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/model"
)

func testManager() *Manager {
	return NewManager(strings.Repeat("k", 32), false)
}

func newContext(cookies []*http.Cookie) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/session", nil)
	for _, cookie := range cookies {
		c.Request.AddCookie(cookie)
	}
	return c, rec
}

func issue(m *Manager, admin *model.Admin) ([]*http.Cookie, string, error) {
	c, rec := newContext(nil)
	token, err := m.Issue(c, admin)
	return rec.Result().Cookies(), token, err
}

func TestIssueAndParseSession(t *testing.T) {
	m := testManager()
	admin := &model.Admin{ID: 7, Username: "admin", UpdatedAt: time.Unix(1700000000, 0)}

	cookies, token, err := issue(m, admin)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if token == "" {
		t.Fatal("expected csrf token")
	}

	c, _ := newContext(cookies)
	claims, err := m.Parse(c)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.AdminID != 7 || claims.AdminUpdatedAt != 1700000000 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseRejectsTamperedSession(t *testing.T) {
	m := testManager()
	admin := &model.Admin{ID: 1, UpdatedAt: time.Unix(1700000000, 0)}
	cookies, _, err := issue(m, admin)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	for i := range cookies {
		if cookies[i].Name != SessionCookie {
			continue
		}
		// 改签名首字符，保证解码后的字节必然不同；末位字符可能只占半个字节，
		// base64 解码会忽略未使用的尾位，改它等于没改。
		parts := strings.SplitN(cookies[i].Value, ".", 2)
		if len(parts) != 2 || parts[1] == "" {
			t.Fatalf("unexpected cookie value: %q", cookies[i].Value)
		}
		head := "A"
		if parts[1][0] == 'A' {
			head = "B"
		}
		cookies[i].Value = parts[0] + "." + head + parts[1][1:]
	}

	c, _ := newContext(cookies)
	if _, err := m.Parse(c); !errors.Is(err, ErrSessionInvalid) {
		t.Fatalf("expected invalid session, got %v", err)
	}
}

func TestParseRejectsExpiredSession(t *testing.T) {
	m := testManager()
	admin := &model.Admin{ID: 1, UpdatedAt: time.Unix(1700000000, 0)}
	cookies, _, err := issue(m, admin)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	m.now = func() time.Time { return time.Now().Add(sessionTTL + time.Minute) }
	c, _ := newContext(cookies)
	if _, err := m.Parse(c); !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("expected expired session, got %v", err)
	}
}

func TestCSRFTokenMustMatch(t *testing.T) {
	m := testManager()
	admin := &model.Admin{ID: 1, UpdatedAt: time.Unix(1700000000, 0)}
	cookies, token, err := issue(m, admin)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	c, _ := newContext(cookies)
	c.Request.Header.Set(CSRFHeader, token)
	if err := m.VerifyCSRF(c); err != nil {
		t.Fatalf("expected csrf pass, got %v", err)
	}

	c2, _ := newContext(cookies)
	c2.Request.Header.Set(CSRFHeader, token+"x")
	if err := m.VerifyCSRF(c2); !errors.Is(err, ErrCSRF) {
		t.Fatalf("expected csrf failure, got %v", err)
	}
}

func TestLoginRateLimit(t *testing.T) {
	m := testManager()
	key := "192.0.2.10"

	for i := 0; i < loginMaxFail; i++ {
		if m.LoginLimited(key) {
			t.Fatalf("blocked too early at %d", i)
		}
		m.RecordLoginFailure(key)
	}
	if !m.LoginLimited(key) {
		t.Fatal("expected blocked after max failures")
	}

	m.ResetLogin(key)
	if m.LoginLimited(key) {
		t.Fatal("expected reset")
	}
}
