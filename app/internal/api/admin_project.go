package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/model"
	"dev-notes-platform/internal/store"
	"dev-notes-platform/internal/validate"
)

type projectInput struct {
	Name      string       `json:"name"`
	Slug      string       `json:"slug"`
	Summary   string       `json:"summary"`
	TechStack []string     `json:"tech_stack"`
	RepoURL   string       `json:"repo_url"`
	SortOrder int          `json:"sort_order"`
	Status    model.Status `json:"status"`
}

func (in *projectInput) toModel(id int64) *model.Project {
	stack := make([]string, 0, len(in.TechStack))
	for _, item := range in.TechStack {
		item = strings.TrimSpace(item)
		if item != "" {
			stack = append(stack, item)
		}
	}
	return &model.Project{
		ID:        id,
		Name:      strings.TrimSpace(in.Name),
		Slug:      strings.TrimSpace(in.Slug),
		Summary:   strings.TrimSpace(in.Summary),
		TechStack: stack,
		RepoURL:   strings.TrimSpace(in.RepoURL),
		SortOrder: in.SortOrder,
		Status:    in.Status,
	}
}

func (s *Server) adminListProjects(c *gin.Context) {
	projects, err := s.store.ListAllProjects(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "项目列表加载失败")
		return
	}
	respondOK(c, gin.H{"items": projects, "total": len(projects)})
}

func (s *Server) adminCreateProject(c *gin.Context) {
	var in projectInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	project := in.toModel(0)
	if fieldErr := validate.Project(project); fieldErr != nil {
		respondFieldError(c, fieldErr)
		return
	}

	id, err := s.store.CreateProject(c.Request.Context(), project)
	if errors.Is(err, store.ErrConflict) {
		respondError(c, http.StatusConflict, "SLUG_CONFLICT", "链接标识已被占用，请换一个")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "项目保存失败")
		return
	}

	created, err := s.store.GetProjectByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "项目已保存，读取失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": created})
}

func (s *Server) adminUpdateProject(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}

	var in projectInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	project := in.toModel(id)
	if fieldErr := validate.Project(project); fieldErr != nil {
		respondFieldError(c, fieldErr)
		return
	}

	if err := s.store.UpdateProject(c.Request.Context(), project); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			respondError(c, http.StatusConflict, "SLUG_CONFLICT", "链接标识已被占用，请换一个")
		case errors.Is(err, store.ErrNotFound):
			respondError(c, http.StatusNotFound, "NOT_FOUND", "项目不存在")
		default:
			respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "项目保存失败")
		}
		return
	}

	updated, err := s.store.GetProjectByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "项目已保存，读取失败")
		return
	}
	respondOK(c, updated)
}

func (s *Server) adminDeleteProject(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := s.store.DeleteProject(c.Request.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "项目不存在")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "项目删除失败")
		return
	}
	respondOK(c, gin.H{"ok": true})
}
