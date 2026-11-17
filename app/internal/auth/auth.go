// Package auth 实现单管理员会话：签名 Cookie、CSRF 双提交与登录失败限流。
//
// 会话不落库：Cookie 里带过期时间与 admin.updated_at，改密后旧 Cookie 自动失效。
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/model"
)

const (
	SessionCookie = "dn_session"
	CSRFCookie    = "dn_csrf"
	CSRFHeader    = "X-CSRF-Token"

	contextAdminKey = "auth.admin"

	sessionTTL   = 12 * time.Hour
	payloadVer   = "v1"
	loginWindow  = 15 * time.Minute
	loginMaxFail = 5
)

var (
	ErrNoSession      = errors.New("会话不存在")
	ErrSessionExpired = errors.New("会话已过期")
	ErrSessionInvalid = errors.New("会话无效")
	ErrCSRF           = errors.New("CSRF 校验失败")
)

// Claims 是会话里携带的信息。
type Claims struct {
	AdminID   int64
	ExpiresAt time.Time
	// AdminUpdatedAt 与 admin.updated_at 比对，密码变更后旧会话失效。
	AdminUpdatedAt int64
}

// Manager 负责会话 cookie、CSRF token 与登录限流。
type Manager struct {
	secret  []byte
	secure  bool
	now     func() time.Time
	limiter *limiter
}

// NewManager 创建会话管理器；secret 来自配置且长度已在启动时校验。
func NewManager(secret string, cookieSecure bool) *Manager {
	return &Manager{
		secret:  []byte(secret),
		secure:  cookieSecure,
		now:     time.Now,
		limiter: newLimiter(loginWindow, loginMaxFail),
	}
}

// Issue 写入会话与 CSRF cookie，并返回给前端使用的 CSRF token。
func (m *Manager) Issue(c *gin.Context, admin *model.Admin) (string, error) {
	expires := m.now().Add(sessionTTL)
	payload := fmt.Sprintf("%s|%d|%d|%d", payloadVer, admin.ID, expires.Unix(), admin.UpdatedAt.Unix())
	value := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + m.sign(payload)

	m.writeCookie(c, SessionCookie, value, true, int(sessionTTL.Seconds()))

	token, err := newToken()
	if err != nil {
		return "", err
	}
	// CSRF token 需要被前端读取后放进请求头，因此不是 HttpOnly。
	m.writeCookie(c, CSRFCookie, token, false, int(sessionTTL.Seconds()))
	return token, nil
}

// Clear 清除会话与 CSRF cookie。
func (m *Manager) Clear(c *gin.Context) {
	m.writeCookie(c, SessionCookie, "", true, -1)
	m.writeCookie(c, CSRFCookie, "", false, -1)
}

func (m *Manager) writeCookie(c *gin.Context, name, value string, httpOnly bool, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", m.secure, httpOnly)
}

// Parse 校验会话 cookie 并返回 claims。
func (m *Manager) Parse(c *gin.Context) (*Claims, error) {
	raw, err := c.Cookie(SessionCookie)
	if err != nil || raw == "" {
		return nil, ErrNoSession
	}

	encoded, signature, ok := strings.Cut(raw, ".")
	if !ok {
		return nil, ErrSessionInvalid
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrSessionInvalid
	}
	payload := string(payloadBytes)
	if !hmac.Equal([]byte(signature), []byte(m.sign(payload))) {
		return nil, ErrSessionInvalid
	}

	parts := strings.Split(payload, "|")
	if len(parts) != 4 || parts[0] != payloadVer {
		return nil, ErrSessionInvalid
	}
	adminID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, ErrSessionInvalid
	}
	expiresUnix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, ErrSessionInvalid
	}
	updatedUnix, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, ErrSessionInvalid
	}

	expiresAt := time.Unix(expiresUnix, 0)
	if m.now().After(expiresAt) {
		return nil, ErrSessionExpired
	}
	return &Claims{AdminID: adminID, ExpiresAt: expiresAt, AdminUpdatedAt: updatedUnix}, nil
}

// CSRFToken 读取当前请求里的 CSRF token，供 /api/admin/me 返回给前端。
func (m *Manager) CSRFToken(c *gin.Context) string {
	token, err := c.Cookie(CSRFCookie)
	if err != nil {
		return ""
	}
	return token
}

// VerifyCSRF 校验双提交 token：请求头与 cookie 必须一致。
func (m *Manager) VerifyCSRF(c *gin.Context) error {
	cookieToken, err := c.Cookie(CSRFCookie)
	if err != nil || cookieToken == "" {
		return ErrCSRF
	}
	headerToken := c.GetHeader(CSRFHeader)
	if headerToken == "" {
		return ErrCSRF
	}
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
		return ErrCSRF
	}
	return nil
}

func (m *Manager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// LoginLimited 判断该来源是否已被限流。
func (m *Manager) LoginLimited(key string) bool {
	return m.limiter.blocked(key)
}

// RecordLoginFailure 记录一次失败；成功登录时调用 ResetLogin。
func (m *Manager) RecordLoginFailure(key string) {
	m.limiter.fail(key)
}

// ResetLogin 清零失败计数。
func (m *Manager) ResetLogin(key string) {
	m.limiter.reset(key)
}

// AdminLoader 按 ID 加载管理员，由调用方注入 store 实现。
type AdminLoader func(ctx context.Context, id int64) (*model.Admin, error)

// RequireAdmin 校验会话并把管理员放入 context。
func (m *Manager) RequireAdmin(load AdminLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.Parse(c)
		if err != nil {
			unauthorized(c, "登录已失效，请重新登录")
			return
		}
		admin, err := load(c.Request.Context(), claims.AdminID)
		if err != nil {
			unauthorized(c, "登录已失效，请重新登录")
			return
		}
		if admin.UpdatedAt.Unix() != claims.AdminUpdatedAt {
			unauthorized(c, "登录已失效，请重新登录")
			return
		}
		c.Set(contextAdminKey, admin)
		c.Next()
	}
}

// CurrentAdmin 取出当前登录管理员。
func CurrentAdmin(c *gin.Context) (*model.Admin, bool) {
	value, ok := c.Get(contextAdminKey)
	if !ok {
		return nil, false
	}
	admin, ok := value.(*model.Admin)
	return admin, ok
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{"code": "UNAUTHORIZED", "message": message},
	})
}

func newToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

// limiter 是进程内的登录失败计数，单管理员场景不需要持久化。
type limiter struct {
	mu       sync.Mutex
	window   time.Duration
	max      int
	attempts map[string]attempt
	now      func() time.Time
}

type attempt struct {
	count   int
	resetAt time.Time
}

func newLimiter(window time.Duration, max int) *limiter {
	return &limiter{
		window:   window,
		max:      max,
		attempts: make(map[string]attempt),
		now:      time.Now,
	}
}

func (l *limiter) blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.attempts[key]
	if !ok {
		return false
	}
	if l.now().After(item.resetAt) {
		delete(l.attempts, key)
		return false
	}
	return item.count >= l.max
}

func (l *limiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	item, ok := l.attempts[key]
	if !ok || now.After(item.resetAt) {
		l.attempts[key] = attempt{count: 1, resetAt: now.Add(l.window)}
		return
	}
	item.count++
	l.attempts[key] = item
}

func (l *limiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}
