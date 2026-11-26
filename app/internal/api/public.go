package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/render"
	"dev-notes-platform/internal/store"
)

func (s *Server) listProjects(c *gin.Context) {
	projects, err := s.store().ListPublishedProjects(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "项目列表加载失败")
		return
	}
	// 列表只给卡片需要的字段，长文说明不必随列表下发。
	for i := range projects {
		projects[i].Description = ""
	}
	respondOK(c, gin.H{"items": projects, "total": len(projects)})
}

func (s *Server) getProject(c *gin.Context) {
	project, err := s.store().GetPublishedProjectBySlug(c.Request.Context(), c.Param("slug"))
	if errors.Is(err, store.ErrNotFound) {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "项目不存在或未公开")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "项目详情加载失败")
		return
	}

	notes, err := s.store().ListPublishedNotes(c.Request.Context(), store.NoteFilter{
		ProjectSlug: project.Slug,
		Limit:       20,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "关联笔记加载失败")
		return
	}
	count, err := s.store().CountPublishedNotesByProject(c.Request.Context(), project.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "关联笔记统计失败")
		return
	}

	// 前台只拿净化后的 HTML，说明源码不进公开响应。
	project.DescriptionHTML = render.Markdown(project.Description)
	project.Description = ""

	respondOK(c, gin.H{
		"project":    project,
		"notes":      notes,
		"note_count": count,
	})
}

func (s *Server) listNotes(c *gin.Context) {
	page := positiveInt(c.Query("page"), 1)
	pageSize := positiveInt(c.Query("page_size"), 10)
	if pageSize > 50 {
		pageSize = 50
	}

	filter := store.NoteFilter{
		ProjectSlug:  c.Query("project"),
		Keyword:      c.Query("q"),
		FeaturedOnly: isTruthy(c.Query("featured")),
		Limit:        pageSize,
		Offset:       (page - 1) * pageSize,
	}
	notes, err := s.store().ListPublishedNotes(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记列表加载失败")
		return
	}
	total, err := s.store().CountPublishedNotes(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记总数统计失败")
		return
	}
	respondList(c, notes, total, page, pageSize)
}

func (s *Server) getNote(c *gin.Context) {
	note, err := s.store().GetPublishedNoteBySlug(c.Request.Context(), c.Param("slug"))
	if errors.Is(err, store.ErrNotFound) {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "笔记不存在或未公开")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记详情加载失败")
		return
	}
	// 前台只拿净化后的 HTML，正文源码不进公开响应。
	note.ContentHTML = render.Markdown(note.Content)
	note.Content = ""
	respondOK(c, note)
}

// isTruthy 把 1 / true / yes 这类查询值当作真，其余按假处理。
func isTruthy(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func positiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
