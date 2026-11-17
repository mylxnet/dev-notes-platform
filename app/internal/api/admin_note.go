package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/model"
	"dev-notes-platform/internal/store"
	"dev-notes-platform/internal/validate"
)

type noteInput struct {
	Title       string       `json:"title"`
	Slug        string       `json:"slug"`
	Summary     string       `json:"summary"`
	Content     string       `json:"content"`
	ProjectID   *int64       `json:"project_id"`
	Status      model.Status `json:"status"`
	PublishedAt string       `json:"published_at"`
}

func (in *noteInput) toModel(id int64) (*model.Note, error) {
	note := &model.Note{
		ID:        id,
		Title:     strings.TrimSpace(in.Title),
		Slug:      strings.TrimSpace(in.Slug),
		Summary:   strings.TrimSpace(in.Summary),
		Content:   in.Content,
		ProjectID: in.ProjectID,
		Status:    in.Status,
	}
	if in.PublishedAt != "" {
		parsed, err := time.Parse(time.RFC3339, in.PublishedAt)
		if err != nil {
			return nil, errors.New("发布时间格式需为 RFC3339")
		}
		utc := parsed.UTC().Truncate(time.Second)
		note.PublishedAt = &utc
	}
	return note, nil
}

func (s *Server) adminListNotes(c *gin.Context) {
	page := positiveInt(c.Query("page"), 1)
	pageSize := positiveInt(c.Query("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	filter := store.AdminNoteFilter{
		Status:  model.Status(strings.TrimSpace(c.Query("status"))),
		Keyword: strings.TrimSpace(c.Query("q")),
		Limit:   pageSize,
		Offset:  (page - 1) * pageSize,
	}

	notes, err := s.store.ListAllNotes(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记列表加载失败")
		return
	}
	total, err := s.store.CountAllNotes(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记总数统计失败")
		return
	}

	// 列表不返回正文，减少响应体积。
	for i := range notes {
		notes[i].Content = ""
	}
	respondList(c, notes, total, page, pageSize)
}

func (s *Server) adminGetNote(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	note, err := s.store.GetNoteByID(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "笔记不存在")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记加载失败")
		return
	}
	respondOK(c, note)
}

func (s *Server) adminCreateNote(c *gin.Context) {
	var in noteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	note, err := in.toModel(0)
	if err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
		return
	}
	if fieldErr := validate.Note(note); fieldErr != nil {
		respondFieldError(c, fieldErr)
		return
	}

	id, err := s.store.CreateNote(c.Request.Context(), note)
	if errors.Is(err, store.ErrConflict) {
		respondError(c, http.StatusConflict, "SLUG_CONFLICT", "链接标识已被占用，请换一个")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记保存失败")
		return
	}

	created, err := s.store.GetNoteByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记已保存，读取失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": created})
}

func (s *Server) adminUpdateNote(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}

	var in noteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	note, err := in.toModel(id)
	if err != nil {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
		return
	}

	existing, err := s.store.GetNoteByID(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respondError(c, http.StatusNotFound, "NOT_FOUND", "笔记不存在")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记加载失败")
		return
	}
	// 已经是发布状态且本次未改发布时间的，沿用原发布时间。
	if note.PublishedAt == nil && existing.PublishedAt != nil {
		note.PublishedAt = existing.PublishedAt
	}

	if fieldErr := validate.Note(note); fieldErr != nil {
		respondFieldError(c, fieldErr)
		return
	}

	if err := s.store.UpdateNote(c.Request.Context(), note); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			respondError(c, http.StatusConflict, "SLUG_CONFLICT", "链接标识已被占用，请换一个")
		case errors.Is(err, store.ErrNotFound):
			respondError(c, http.StatusNotFound, "NOT_FOUND", "笔记不存在")
		default:
			respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记保存失败")
		}
		return
	}

	updated, err := s.store.GetNoteByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记已保存，读取失败")
		return
	}
	respondOK(c, updated)
}

func (s *Server) adminDeleteNote(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := s.store.DeleteNote(c.Request.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "笔记不存在")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "笔记删除失败")
		return
	}
	respondOK(c, gin.H{"ok": true})
}

func pathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "记录 ID 不合法")
		return 0, false
	}
	return id, true
}
