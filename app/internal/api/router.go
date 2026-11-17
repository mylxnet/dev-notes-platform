package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/auth"
	"dev-notes-platform/internal/model"
	"dev-notes-platform/internal/store"
	"dev-notes-platform/internal/webui"
)

// 管理接口请求体上限 1 MB。
const maxAdminBody = 1 << 20

// Server 持有 API 依赖。
type Server struct {
	store   *store.Store
	auth    *auth.Manager
	version string
}

// New 构建 Gin 路由。
func New(st *store.Store, version string, sessions *auth.Manager) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), requestID(), securityHeaders())

	server := &Server{store: st, auth: sessions, version: version}

	router.GET("/healthz", server.health)

	api := router.Group("/api")
	api.GET("/projects", server.listProjects)
	api.GET("/projects/:slug", server.getProject)
	api.GET("/notes", server.listNotes)
	api.GET("/notes/:slug", server.getNote)

	// 登录不需要会话，也不校验 CSRF。
	api.POST("/admin/login", bodyLimit(maxAdminBody), server.adminLogin)

	loadAdmin := func(ctx context.Context, id int64) (*model.Admin, error) {
		return st.GetAdminByID(ctx, id)
	}
	admin := api.Group("/admin",
		bodyLimit(maxAdminBody),
		server.auth.RequireAdmin(loadAdmin),
		server.csrfGuard(),
	)
	admin.POST("/logout", server.adminLogout)
	admin.GET("/me", server.adminMe)
	admin.PUT("/password", server.adminChangePassword)
	admin.POST("/preview", server.adminPreview)

	admin.GET("/notes", server.adminListNotes)
	admin.GET("/notes/:id", server.adminGetNote)
	admin.POST("/notes", server.adminCreateNote)
	admin.PUT("/notes/:id", server.adminUpdateNote)
	admin.DELETE("/notes/:id", server.adminDeleteNote)

	admin.GET("/projects", server.adminListProjects)
	admin.POST("/projects", server.adminCreateProject)
	admin.PUT("/projects/:id", server.adminUpdateProject)
	admin.DELETE("/projects/:id", server.adminDeleteProject)

	router.NoRoute(noRouteHandler(webui.FS()))

	return router
}

// csrfGuard 校验管理端写操作的双提交 token。
func (s *Server) csrfGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		if err := s.auth.VerifyCSRF(c); err != nil {
			respondError(c, http.StatusForbidden, "CSRF_FAILED", "请求校验失败，请刷新页面后重试")
			return
		}
		c.Next()
	}
}

// bodyLimit 拒绝超限请求体，避免大字段拖垮服务。
func bodyLimit(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > max {
			respondError(c, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "请求内容过大")
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		}
		c.Next()
	}
}

// noRouteHandler 处理接口 404、静态资源与 SPA 回退。
// 前端使用 history 路由，/notes/:slug、/admin/notes 这类真实路径需要回退到 index.html。
func noRouteHandler(staticFS fs.FS) gin.HandlerFunc {
	var fileServer http.Handler
	var indexHTML []byte
	if staticFS != nil {
		fileServer = http.FileServer(http.FS(staticFS))
		indexHTML, _ = fs.ReadFile(staticFS, "index.html")
	}

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "接口不存在")
			return
		}

		if staticFS != nil {
			requestPath := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
			if requestPath != "" && requestPath != "." {
				if info, err := fs.Stat(staticFS, requestPath); err == nil && !info.IsDir() {
					// 构建产物带内容哈希，可长缓存。
					if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
						c.Header("Cache-Control", "public, max-age=31536000, immutable")
					}
					fileServer.ServeHTTP(c.Writer, c.Request)
					return
				}
			}
			if len(indexHTML) > 0 {
				c.Header("Cache-Control", "no-cache")
				c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
				return
			}
		}

		c.String(http.StatusNotFound, "not found")
	}
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			var raw [12]byte
			if _, err := rand.Read(raw[:]); err == nil {
				id = hex.EncodeToString(raw[:])
			} else {
				id = time.Now().UTC().Format("20060102150405.000000000")
			}
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func securityHeaders() gin.HandlerFunc {
	const policy = "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; " +
		"script-src 'self'; base-uri 'self'; frame-ancestors 'none'"
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", policy)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
