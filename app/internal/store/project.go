package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"dev-notes-platform/internal/model"
)

// ListPublishedProjects 返回全部已发布项目。
func (s *Store) ListPublishedProjects(ctx context.Context) ([]model.Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+projectColumns+`
		FROM project
		WHERE status = ?
		ORDER BY sort_order ASC, id ASC`, model.StatusPublished)
	if err != nil {
		return nil, fmt.Errorf("list published projects: %w", err)
	}
	defer rows.Close()

	projects := make([]model.Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, *project)
	}
	return projects, rows.Err()
}

// AdminProjectFilter 是后台项目列表的分页参数。
type AdminProjectFilter struct {
	Limit  int
	Offset int
}

// ListAllProjects 返回全部状态的项目，供后台使用；按排序值与 ID 稳定分页。
func (s *Store) ListAllProjects(ctx context.Context, filter AdminProjectFilter) ([]model.Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+projectColumns+`
		FROM project
		ORDER BY sort_order ASC, id ASC
		LIMIT ? OFFSET ?`,
		normalizeLimit(filter.Limit), normalizeOffset(filter.Offset))
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]model.Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, *project)
	}
	return projects, rows.Err()
}

// CountAllProjects 返回全部状态的项目总数，供后台分页使用。
func (s *Store) CountAllProjects(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count projects: %w", err)
	}
	return count, nil
}

// GetPublishedProjectBySlug 返回已发布项目详情。
func (s *Store) GetPublishedProjectBySlug(ctx context.Context, slug string) (*model.Project, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+projectColumns+`
		FROM project WHERE slug = ? AND status = ?`, slug, model.StatusPublished)
	project, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get project by slug: %w", err)
	}
	return project, nil
}

// GetProjectByID 返回任意状态的项目。
func (s *Store) GetProjectByID(ctx context.Context, id int64) (*model.Project, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+projectColumns+` FROM project WHERE id = ?`, id)
	project, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}
	return project, nil
}

// CreateProject 新建项目并返回 ID。
func (s *Store) CreateProject(ctx context.Context, project *model.Project) (int64, error) {
	if project.TechStack == nil {
		project.TechStack = []string{}
	}
	techStack, err := json.Marshal(project.TechStack)
	if err != nil {
		return 0, fmt.Errorf("marshal tech stack: %w", err)
	}
	now := nowText()
	result, err := s.db.ExecContext(ctx, `INSERT INTO project
		(name, slug, summary, tech_stack, repo_url, description, cover_image, sort_order, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		project.Name,
		project.Slug,
		project.Summary,
		string(techStack),
		project.RepoURL,
		project.Description,
		project.CoverImage,
		project.SortOrder,
		project.Status,
		now,
		now,
	)
	if err != nil {
		if isUniqueError(err) {
			return 0, ErrConflict
		}
		return 0, fmt.Errorf("create project: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("project id: %w", err)
	}
	project.ID = id
	project.CreatedAt = parseTime(now)
	project.UpdatedAt = parseTime(now)
	return id, nil
}

// UpdateProject 更新项目。
func (s *Store) UpdateProject(ctx context.Context, project *model.Project) error {
	if project.TechStack == nil {
		project.TechStack = []string{}
	}
	techStack, err := json.Marshal(project.TechStack)
	if err != nil {
		return fmt.Errorf("marshal tech stack: %w", err)
	}
	now := nowText()
	result, err := s.db.ExecContext(ctx, `UPDATE project SET
		name = ?, slug = ?, summary = ?, tech_stack = ?, repo_url = ?,
		description = ?, cover_image = ?, sort_order = ?, status = ?, updated_at = ?
		WHERE id = ?`,
		project.Name,
		project.Slug,
		project.Summary,
		string(techStack),
		project.RepoURL,
		project.Description,
		project.CoverImage,
		project.SortOrder,
		project.Status,
		now,
		project.ID,
	)
	if err != nil {
		if isUniqueError(err) {
			return ErrConflict
		}
		return fmt.Errorf("update project: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("project rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	project.UpdatedAt = parseTime(now)
	return nil
}

// DeleteProject 删除项目；关联笔记由外键置空。
func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM project WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("project rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountPublishedNotesByProject 统计项目下已发布笔记数量。
func (s *Store) CountPublishedNotesByProject(ctx context.Context, projectID int64) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM note
		WHERE project_id = ? AND status = ?`, projectID, model.StatusPublished).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count project notes: %w", err)
	}
	return count, nil
}
