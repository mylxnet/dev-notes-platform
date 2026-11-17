package api

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/store"
)

// Server 持有 API 依赖。
type Server struct {
	store   *store.Store
	version string
}

// New 构建 Gin 路由。
func New(st *store.Store, version string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), requestID(), securityHeaders())

	server := &Server{store: st, version: version}

	router.GET("/healthz", server.health)

	api := router.Group("/api")
	api.GET("/projects", server.listProjects)
	api.GET("/projects/:slug", server.getProject)
	api.GET("/notes", server.listNotes)
	api.GET("/notes/:slug", server.getNote)

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			respondError(c, 404, "NOT_FOUND", "接口不存在")
			return
		}
		c.String(404, "not found")
	})

	return router
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
