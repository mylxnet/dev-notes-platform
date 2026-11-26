package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"dev-notes-platform/internal/auth"
	"dev-notes-platform/internal/model"
	"dev-notes-platform/internal/render"
	"dev-notes-platform/internal/store"
	"dev-notes-platform/internal/validate"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type passwordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type previewRequest struct {
	Content string `json:"content"`
}

func (s *Server) adminLogin(c *gin.Context) {
	clientKey := c.ClientIP()
	if s.auth.LoginLimited(clientKey) {
		respondError(c, http.StatusTooManyRequests, "TOO_MANY_ATTEMPTS", "登录失败次数过多，请 15 分钟后再试")
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "请输入账号与密码")
		return
	}

	admin, err := s.store().GetAdminByUsername(c.Request.Context(), req.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) != nil {
		s.auth.RecordLoginFailure(clientKey)
		// 不区分「账号不存在」与「密码错误」。
		respondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "账号或密码不正确")
		return
	}

	s.auth.ResetLogin(clientKey)
	token, err := s.auth.Issue(c, admin)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "会话创建失败")
		return
	}
	respondOK(c, gin.H{"admin": adminView(admin), "csrf_token": token})
}

func (s *Server) adminLogout(c *gin.Context) {
	s.auth.Clear(c)
	respondOK(c, gin.H{"ok": true})
}

func (s *Server) adminMe(c *gin.Context) {
	admin, ok := auth.CurrentAdmin(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "登录已失效，请重新登录")
		return
	}
	respondOK(c, gin.H{"admin": adminView(admin), "csrf_token": s.auth.CSRFToken(c)})
}

func (s *Server) adminChangePassword(c *gin.Context) {
	admin, ok := auth.CurrentAdmin(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "登录已失效，请重新登录")
		return
	}

	var req passwordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.CurrentPassword)) != nil {
		respondError(c, http.StatusUnprocessableEntity, "INVALID_PASSWORD", "当前密码不正确")
		return
	}
	if err := validate.Password(req.NewPassword); err != nil {
		respondFieldError(c, err)
		return
	}

	if err := s.store().UpdateAdminPassword(c.Request.Context(), admin.ID, req.NewPassword); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "管理员不存在")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "密码更新失败")
		return
	}

	// 改密后 updated_at 变化会让旧会话失效，这里立即签发新会话，避免当前页面被踢出。
	updated, err := s.store().GetAdminByUsername(c.Request.Context(), admin.Username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "密码已更新，请重新登录")
		return
	}
	token, err := s.auth.Issue(c, updated)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "密码已更新，请重新登录")
		return
	}
	respondOK(c, gin.H{"admin": adminView(updated), "csrf_token": token})
}

func (s *Server) adminPreview(c *gin.Context) {
	var req previewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	respondOK(c, gin.H{"html": render.Markdown(req.Content)})
}

// adminView 只暴露后台需要的字段。
func adminView(admin *model.Admin) gin.H {
	return gin.H{
		"id":         admin.ID,
		"username":   admin.Username,
		"updated_at": admin.UpdatedAt,
	}
}

// respondFieldError 把校验失败转成 422；命中词汇时使用 CONTENT_BLOCKED。
func respondFieldError(c *gin.Context, err *validate.FieldError) {
	code := "VALIDATION_FAILED"
	if len(err.Hits) > 0 {
		code = "CONTENT_BLOCKED"
	}
	body := gin.H{
		"code":    code,
		"message": err.Message,
		"field":   err.Field,
	}
	if len(err.Hits) > 0 {
		body["hits"] = err.Hits
	}
	if err.Suggestion != "" {
		body["suggestion"] = err.Suggestion
	}
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": body})
}
