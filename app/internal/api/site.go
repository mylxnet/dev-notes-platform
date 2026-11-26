package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/validate"
)

const (
	// settingICPNumber 是站点配置表里备案号对应的键。
	settingICPNumber = "icp_number"
	// maxFeaturedNotes 是首页推荐面板的展示上限，与前台面板槽位一致；全量推荐走笔记列表。
	maxFeaturedNotes = 3
)

type settingsInput struct {
	ICPNumber string `json:"icp_number"`
}

// featuredNoteView 是首页推荐面板需要的最小字段集合。
type featuredNoteView struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
}

// featuredNotes 取被标记为推荐且已发布的笔记；推荐是笔记自身属性，不再由站点设置维护。
func (s *Server) featuredNotes(ctx context.Context) []featuredNoteView {
	notes, err := s.store.ListFeaturedNotes(ctx, maxFeaturedNotes)
	if err != nil {
		return []featuredNoteView{}
	}
	views := make([]featuredNoteView, 0, len(notes))
	for _, note := range notes {
		publishedAt := ""
		if note.PublishedAt != nil {
			publishedAt = note.PublishedAt.Format(time.RFC3339)
		}
		views = append(views, featuredNoteView{Slug: note.Slug, Title: note.Title, PublishedAt: publishedAt})
	}
	return views
}

// siteInfo 返回前台页脚需要的备案号与首页推荐笔记；未配置时返回空值。
func (s *Server) siteInfo(c *gin.Context) {
	ctx := c.Request.Context()
	icp, err := s.store.GetSetting(ctx, settingICPNumber)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点信息加载失败")
		return
	}
	respondOK(c, gin.H{"icp_number": icp, "featured_notes": s.featuredNotes(ctx)})
}

// adminGetSettings 返回后台可编辑的站点配置（只剩备案号，推荐已迁到笔记表单）。
func (s *Server) adminGetSettings(c *gin.Context) {
	icp, err := s.store.GetSetting(c.Request.Context(), settingICPNumber)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点设置加载失败")
		return
	}
	respondOK(c, gin.H{"icp_number": icp})
}

// adminUpdateSettings 保存站点配置；清空备案号后前台恢复占位文案。
func (s *Server) adminUpdateSettings(c *gin.Context) {
	ctx := c.Request.Context()
	var in settingsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	in.ICPNumber = strings.TrimSpace(in.ICPNumber)
	if fieldErr := validate.ICPNumber(in.ICPNumber); fieldErr != nil {
		respondFieldError(c, fieldErr)
		return
	}
	if err := s.store.SetSetting(ctx, settingICPNumber, in.ICPNumber); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点设置保存失败")
		return
	}
	respondOK(c, gin.H{"icp_number": in.ICPNumber})
}
