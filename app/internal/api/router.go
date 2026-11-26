package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/auth"
	"dev-notes-platform/internal/backup"
	"dev-notes-platform/internal/model"
	"dev-notes-platform/internal/store"
	"dev-notes-platform/internal/webui"
)

// 管理接口请求体上限 1 MB。
const maxAdminBody = 1 << 20

// Server 持有 API 依赖。
type Server struct {
	db       *storeRef
	auth     *auth.Manager
	version  string
	dataDir  string
	dbPath   string
	mediaDir string
	// restoreMu 保证同一时间只有一个恢复流程在跑。
	restoreMu sync.Mutex
	// maintenance 为真时拒绝其余 API 请求，避免恢复替换文件期间读到半个库。
	maintenance atomic.Bool
}

// storeRef 让「从备份恢复」可以在不重启进程的前提下替换数据库连接。
type storeRef struct {
	mu sync.RWMutex
	st *store.Store
}

func newStoreRef(st *store.Store) *storeRef { return &storeRef{st: st} }

func (r *storeRef) get() *store.Store {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.st
}

func (r *storeRef) swap(next *store.Store) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.st = next
}

// store 返回当前数据库实例；恢复流程替换后，后续请求自动拿到新连接。
func (s *Server) store() *store.Store { return s.db.get() }

// New 构建 Gin 路由；dataDir 是数据目录（含 media/），dbPath 是数据库文件位置。
func New(st *store.Store, version string, sessions *auth.Manager, dataDir, dbPath string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), requestID(), securityHeaders())

	server := &Server{
		db:       newStoreRef(st),
		auth:     sessions,
		version:  version,
		dataDir:  dataDir,
		dbPath:   dbPath,
		mediaDir: filepath.Join(dataDir, "media"),
	}

	router.GET("/healthz", server.health)
	router.GET("/media/uploads/:name", server.serveUploadedMedia)

	api := router.Group("/api", server.maintenanceGuard())
	api.GET("/projects", server.listProjects)
	api.GET("/projects/:slug", server.getProject)
	api.GET("/notes", server.listNotes)
	api.GET("/notes/:slug", server.getNote)
	api.GET("/site", server.siteInfo)

	// 登录不需要会话，也不校验 CSRF。
	api.POST("/admin/login", bodyLimit(maxAdminBody), server.adminLogin)

	// 每次请求都取当前连接，避免恢复替换后仍用已关闭的旧库。
	loadAdmin := func(ctx context.Context, id int64) (*model.Admin, error) {
		return server.store().GetAdminByID(ctx, id)
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

	admin.GET("/settings", server.adminGetSettings)
	admin.PUT("/settings", server.adminUpdateSettings)

	admin.GET("/backup/summary", server.adminBackupSummary)
	admin.POST("/backup", server.adminCreateBackup)
	admin.GET("/backup/snapshots", server.adminListSnapshots)
	admin.POST("/backup/snapshots/:name/restore", server.adminRestoreSnapshot)

	// 备份包上传走独立的请求体上限（包本身限制在 backup.MaxArchiveBytes）。
	backupGroup := api.Group("/admin",
		server.auth.RequireAdmin(loadAdmin),
		server.csrfGuard(),
		bodyLimit(backup.MaxArchiveBytes+(8<<20)),
	)
	backupGroup.POST("/backup/inspect", server.adminInspectBackup)
	backupGroup.POST("/backup/restore", server.adminRestoreBackup)

	// 上传单独放行更大的请求体（文件本身仍限制 2 MB），鉴权与 CSRF 与管理端一致。
	upload := api.Group("/admin",
		server.auth.RequireAdmin(loadAdmin),
		server.csrfGuard(),
		bodyLimit(3<<20),
	)
	upload.POST("/media/upload", server.adminUploadMedia)

	router.NoRoute(noRouteHandler(webui.FS()))

	return router
}

// maintenanceGuard 在恢复数据期间挡住其余接口，避免读到替换中的数据库。
// 备份接口自身放行（恢复流程需要它们先完成校验与会话）。
func (s *Server) maintenanceGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.maintenance.Load() {
			c.Next()
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/admin/backup") {
			c.Next()
			return
		}
		c.Header("Retry-After", "3")
		respondError(c, http.StatusServiceUnavailable, "MAINTENANCE", "正在恢复数据，请稍候重试")
	}
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
