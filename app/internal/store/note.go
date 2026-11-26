package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"dev-notes-platform/internal/model"
)

// NoteFilter 是前台笔记列表的筛选条件。
type NoteFilter struct {
	ProjectSlug string
	Keyword     string
	// FeaturedOnly 只看被标记为推荐的笔记（首页推荐面板的「全部」用）。
	FeaturedOnly bool
	Limit        int
	Offset       int
}

// AdminNoteFilter 是后台笔记列表的筛选条件。
type AdminNoteFilter struct {
	Status  model.Status
	Keyword string
	Limit   int
	Offset  int
}

// ListPublishedNotes 返回已发布笔记列表。
func (s *Store) ListPublishedNotes(ctx context.Context, filter NoteFilter) ([]model.Note, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT ` + noteColumns + `
		FROM note n
		LEFT JOIN project p ON p.id = n.project_id
		WHERE n.status = ?`)
	args := []any{model.StatusPublished}

	if filter.ProjectSlug != "" {
		builder.WriteString(` AND p.slug = ?`)
		args = append(args, filter.ProjectSlug)
	}
	if filter.Keyword != "" {
		builder.WriteString(` AND (n.title LIKE ? OR n.summary LIKE ?)`)
		like := "%" + filter.Keyword + "%"
		args = append(args, like, like)
	}
	if filter.FeaturedOnly {
		builder.WriteString(` AND n.is_featured = 1`)
	}
	builder.WriteString(` ORDER BY n.published_at DESC, n.id DESC LIMIT ? OFFSET ?`)
	args = append(args, normalizeLimit(filter.Limit), normalizeOffset(filter.Offset))

	rows, err := s.db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list published notes: %w", err)
	}
	defer rows.Close()

	notes := make([]model.Note, 0)
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, *note)
	}
	return notes, rows.Err()
}

// ListFeaturedNotes 返回首页推荐面板用的已发布推荐笔记，按发布时间倒序取前 limit 条。
func (s *Store) ListFeaturedNotes(ctx context.Context, limit int) ([]model.Note, error) {
	return s.ListPublishedNotes(ctx, NoteFilter{FeaturedOnly: true, Limit: limit})
}

// CountPublishedNotes 返回已发布笔记总数。
func (s *Store) CountPublishedNotes(ctx context.Context, filter NoteFilter) (int, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT COUNT(*)
		FROM note n
		LEFT JOIN project p ON p.id = n.project_id
		WHERE n.status = ?`)
	args := []any{model.StatusPublished}

	if filter.ProjectSlug != "" {
		builder.WriteString(` AND p.slug = ?`)
		args = append(args, filter.ProjectSlug)
	}
	if filter.Keyword != "" {
		builder.WriteString(` AND (n.title LIKE ? OR n.summary LIKE ?)`)
		like := "%" + filter.Keyword + "%"
		args = append(args, like, like)
	}
	if filter.FeaturedOnly {
		builder.WriteString(` AND n.is_featured = 1`)
	}

	var count int
	if err := s.db.QueryRowContext(ctx, builder.String(), args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count published notes: %w", err)
	}
	return count, nil
}

// GetPublishedNoteBySlug 返回已发布笔记详情。
func (s *Store) GetPublishedNoteBySlug(ctx context.Context, slug string) (*model.Note, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+noteColumns+`
		FROM note n
		LEFT JOIN project p ON p.id = n.project_id
		WHERE n.slug = ? AND n.status = ?`, slug, model.StatusPublished)
	note, err := scanNote(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get note by slug: %w", err)
	}
	return note, nil
}

// ListAllNotes 返回后台笔记列表。
func (s *Store) ListAllNotes(ctx context.Context, filter AdminNoteFilter) ([]model.Note, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT ` + noteColumns + `
		FROM note n
		LEFT JOIN project p ON p.id = n.project_id
		WHERE 1 = 1`)
	args := make([]any, 0)

	if filter.Status != "" {
		builder.WriteString(` AND n.status = ?`)
		args = append(args, filter.Status)
	}
	if filter.Keyword != "" {
		builder.WriteString(` AND (n.title LIKE ? OR n.summary LIKE ?)`)
		like := "%" + filter.Keyword + "%"
		args = append(args, like, like)
	}
	builder.WriteString(` ORDER BY n.updated_at DESC, n.id DESC LIMIT ? OFFSET ?`)
	args = append(args, normalizeLimit(filter.Limit), normalizeOffset(filter.Offset))

	rows, err := s.db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]model.Note, 0)
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, *note)
	}
	return notes, rows.Err()
}

// CountAllNotes 返回后台笔记列表总数。
func (s *Store) CountAllNotes(ctx context.Context, filter AdminNoteFilter) (int, error) {
	var builder strings.Builder
	builder.WriteString(`SELECT COUNT(*) FROM note n WHERE 1 = 1`)
	args := make([]any, 0)

	if filter.Status != "" {
		builder.WriteString(` AND n.status = ?`)
		args = append(args, filter.Status)
	}
	if filter.Keyword != "" {
		builder.WriteString(` AND (n.title LIKE ? OR n.summary LIKE ?)`)
		like := "%" + filter.Keyword + "%"
		args = append(args, like, like)
	}

	var count int
	if err := s.db.QueryRowContext(ctx, builder.String(), args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count notes: %w", err)
	}
	return count, nil
}

// GetNoteByID 返回任意状态的笔记。
func (s *Store) GetNoteByID(ctx context.Context, id int64) (*model.Note, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+noteColumns+`
		FROM note n
		LEFT JOIN project p ON p.id = n.project_id
		WHERE n.id = ?`, id)
	note, err := scanNote(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get note by id: %w", err)
	}
	return note, nil
}

// CreateNote 新建笔记并返回 ID。
func (s *Store) CreateNote(ctx context.Context, note *model.Note) (int64, error) {
	if note.PublishedAt == nil && note.Status == model.StatusPublished {
		now := time.Now().UTC().Truncate(time.Second)
		note.PublishedAt = &now
	}
	now := nowText()
	result, err := s.db.ExecContext(ctx, `INSERT INTO note
		(title, slug, summary, content, project_id, is_featured, status, published_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		note.Title,
		note.Slug,
		note.Summary,
		note.Content,
		nullableID(note.ProjectID),
		boolInt(note.IsFeatured),
		note.Status,
		nullableTime(note.PublishedAt),
		now,
		now,
	)
	if err != nil {
		if isUniqueError(err) {
			return 0, ErrConflict
		}
		return 0, fmt.Errorf("create note: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("note id: %w", err)
	}
	note.ID = id
	note.CreatedAt = parseTime(now)
	note.UpdatedAt = parseTime(now)
	return id, nil
}

// UpdateNote 更新笔记。
func (s *Store) UpdateNote(ctx context.Context, note *model.Note) error {
	if note.PublishedAt == nil && note.Status == model.StatusPublished {
		now := time.Now().UTC().Truncate(time.Second)
		note.PublishedAt = &now
	}
	now := nowText()
	result, err := s.db.ExecContext(ctx, `UPDATE note SET
		title = ?, slug = ?, summary = ?, content = ?, project_id = ?, is_featured = ?,
		status = ?, published_at = ?, updated_at = ?
		WHERE id = ?`,
		note.Title,
		note.Slug,
		note.Summary,
		note.Content,
		nullableID(note.ProjectID),
		boolInt(note.IsFeatured),
		note.Status,
		nullableTime(note.PublishedAt),
		now,
		note.ID,
	)
	if err != nil {
		if isUniqueError(err) {
			return ErrConflict
		}
		return fmt.Errorf("update note: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("note rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	note.UpdatedAt = parseTime(now)
	return nil
}

// DeleteNote 删除笔记。
func (s *Store) DeleteNote(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM note WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("note rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
