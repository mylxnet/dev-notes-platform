package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/validate"
)

const (
	// settingICPNumber 是站点配置表里备案号对应的键。
	settingICPNumber = "icp_number"
	// settingFeaturedNotes 存首页推荐笔记的 slug 数组（JSON），顺序即展示顺序。
	settingFeaturedNotes = "featured_notes"
	// maxFeaturedNotes 是首页推荐位上限，与前台面板槽位一致。
	maxFeaturedNotes = 3
)

type settingsInput struct {
	ICPNumber string `json:"icp_number"`
	// FeaturedSlugs 是首页推荐位，按顺序展示，最多 3 个。
	FeaturedSlugs []string `json:"featured_slugs"`
}

// featuredNoteView 是首页推荐面板需要的最小字段集合。
type featuredNoteView struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
}

// readFeaturedSlugs 读取已保存的推荐位：解析失败按未配置处理，最多返回 3 个。
func (s *Server) readFeaturedSlugs(ctx context.Context) []string {
	raw, err := s.store.GetSetting(ctx, settingFeaturedNotes)
	if err != nil || raw == "" {
		return []string{}
	}
	var saved []string
	if err := json.Unmarshal([]byte(raw), &saved); err != nil {
		return []string{}
	}
	slugs := make([]string, 0, maxFeaturedNotes)
	seen := make(map[string]bool, len(saved))
	for _, slug := range saved {
		slug = strings.TrimSpace(slug)
		if slug == "" || seen[slug] || len(slugs) >= maxFeaturedNotes {
			continue
		}
		seen[slug] = true
		slugs = append(slugs, slug)
	}
	return slugs
}

// featuredNotes 把推荐位解析成已发布笔记；已删除或已下线的自动跳过，不报错。
func (s *Server) featuredNotes(ctx context.Context) []featuredNoteView {
	slugs := s.readFeaturedSlugs(ctx)
	notes := make([]featuredNoteView, 0, len(slugs))
	for _, slug := range slugs {
		note, err := s.store.GetPublishedNoteBySlug(ctx, slug)
		if err != nil {
			continue
		}
		publishedAt := ""
		if note.PublishedAt != nil {
			publishedAt = note.PublishedAt.Format(time.RFC3339)
		}
		notes = append(notes, featuredNoteView{Slug: note.Slug, Title: note.Title, PublishedAt: publishedAt})
	}
	return notes
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

// adminGetSettings 返回后台可编辑的站点配置。
func (s *Server) adminGetSettings(c *gin.Context) {
	ctx := c.Request.Context()
	icp, err := s.store.GetSetting(ctx, settingICPNumber)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点设置加载失败")
		return
	}
	respondOK(c, gin.H{"icp_number": icp, "featured_slugs": s.readFeaturedSlugs(ctx)})
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
	slugs, fieldErr := s.normalizeFeaturedSlugs(ctx, in.FeaturedSlugs)
	if fieldErr != nil {
		respondFieldError(c, fieldErr)
		return
	}
	if err := s.store.SetSetting(ctx, settingICPNumber, in.ICPNumber); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点设置保存失败")
		return
	}
	encoded, err := json.Marshal(slugs)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点设置保存失败")
		return
	}
	if err := s.store.SetSetting(ctx, settingFeaturedNotes, string(encoded)); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点设置保存失败")
		return
	}
	respondOK(c, gin.H{"icp_number": in.ICPNumber, "featured_slugs": slugs})
}

// normalizeFeaturedSlugs 校验推荐位：去空去重、最多 3 个、必须存在且已发布。
func (s *Server) normalizeFeaturedSlugs(ctx context.Context, input []string) ([]string, *validate.FieldError) {
	slugs := make([]string, 0, maxFeaturedNotes)
	seen := make(map[string]bool, len(input))
	for _, raw := range input {
		slug := strings.TrimSpace(raw)
		if slug == "" || seen[slug] {
			continue
		}
		if len(slugs) >= maxFeaturedNotes {
			return nil, &validate.FieldError{Field: "featured_slugs", Message: "首页推荐位最多 3 篇"}
		}
		if fieldErr := validate.Slug(slug); fieldErr != nil {
			return nil, &validate.FieldError{Field: "featured_slugs", Message: "推荐笔记的链接标识格式不正确"}
		}
		if _, err := s.store.GetPublishedNoteBySlug(ctx, slug); err != nil {
			return nil, &validate.FieldError{
				Field:      "featured_slugs",
				Message:    fmt.Sprintf("推荐笔记「%s」不存在或未发布，请重新选择", slug),
				Suggestion: "只能从已发布的笔记中选择推荐位",
			}
		}
		seen[slug] = true
		slugs = append(slugs, slug)
	}
	return slugs, nil
}
