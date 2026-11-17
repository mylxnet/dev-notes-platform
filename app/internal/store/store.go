package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"dev-notes-platform/internal/model"

	_ "modernc.org/sqlite"
)

var (
	// ErrNotFound 表示目标记录不存在。
	ErrNotFound = errors.New("record not found")
	// ErrConflict 表示唯一约束冲突。
	ErrConflict = errors.New("record conflict")
)

const schemaVersion = 1

// Store 封装 SQLite 连接与迁移。
type Store struct {
	db *sql.DB
}

// Open 打开数据库、设置 PRAGMA 并执行迁移。
func Open(ctx context.Context, path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("db path is empty")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
	} {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply pragma: %w", err)
		}
	}

	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	return s.db.Close()
}

// Ping 检查数据库可用性。
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) migrate(ctx context.Context) error {
	var current int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if current >= schemaVersion {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	defer tx.Rollback()

	for _, statement := range schemaStatements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", schemaVersion)); err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}
	return nil
}

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS project (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT NOT NULL,
		slug        TEXT NOT NULL UNIQUE,
		summary     TEXT NOT NULL,
		tech_stack  TEXT NOT NULL DEFAULT '[]',
		repo_url    TEXT NOT NULL,
		sort_order  INTEGER NOT NULL DEFAULT 0,
		status      TEXT NOT NULL DEFAULT 'draft'
		            CHECK (status IN ('published', 'draft', 'hidden')),
		created_at  TEXT NOT NULL,
		updated_at  TEXT NOT NULL
	);`,
	`CREATE INDEX IF NOT EXISTS idx_project_status_sort
		ON project (status, sort_order, id);`,
	`CREATE TABLE IF NOT EXISTS note (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		title         TEXT NOT NULL,
		slug          TEXT NOT NULL UNIQUE,
		summary       TEXT NOT NULL,
		content       TEXT NOT NULL,
		project_id    INTEGER,
		status        TEXT NOT NULL DEFAULT 'draft'
		              CHECK (status IN ('published', 'draft', 'hidden')),
		published_at  TEXT,
		created_at    TEXT NOT NULL,
		updated_at    TEXT NOT NULL,
		FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE SET NULL
	);`,
	`CREATE INDEX IF NOT EXISTS idx_note_status_published
		ON note (status, published_at DESC, id DESC);`,
	`CREATE INDEX IF NOT EXISTS idx_note_project
		ON note (project_id, status);`,
	`CREATE TABLE IF NOT EXISTS admin (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		username      TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		updated_at    TEXT NOT NULL
	);`,
}

const projectColumns = `id, name, slug, summary, tech_stack, repo_url, sort_order, status, created_at, updated_at`

const noteColumns = `n.id, n.title, n.slug, n.summary, n.content, n.project_id, p.name, p.slug, n.status, n.published_at, n.created_at, n.updated_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProject(scanner rowScanner) (*model.Project, error) {
	var project model.Project
	var techStack, createdAt, updatedAt string
	if err := scanner.Scan(
		&project.ID,
		&project.Name,
		&project.Slug,
		&project.Summary,
		&techStack,
		&project.RepoURL,
		&project.SortOrder,
		&project.Status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}
	if techStack == "" {
		techStack = "[]"
	}
	if err := json.Unmarshal([]byte(techStack), &project.TechStack); err != nil {
		project.TechStack = []string{}
	}
	project.CreatedAt = parseTime(createdAt)
	project.UpdatedAt = parseTime(updatedAt)
	return &project, nil
}

func scanNote(scanner rowScanner) (*model.Note, error) {
	var note model.Note
	var projectID sql.NullInt64
	var projectName, projectSlug sql.NullString
	var publishedAt sql.NullString
	var createdAt, updatedAt string
	if err := scanner.Scan(
		&note.ID,
		&note.Title,
		&note.Slug,
		&note.Summary,
		&note.Content,
		&projectID,
		&projectName,
		&projectSlug,
		&note.Status,
		&publishedAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}
	if projectID.Valid {
		id := projectID.Int64
		note.ProjectID = &id
	}
	if projectName.Valid && projectSlug.Valid {
		note.Project = &model.ProjectSummary{Name: projectName.String, Slug: projectSlug.String}
	}
	note.PublishedAt = parseNullTime(publishedAt)
	note.CreatedAt = parseTime(createdAt)
	note.UpdatedAt = parseTime(updatedAt)
	return &note, nil
}

func nowText() string {
	return time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
}

func parseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func parseNullTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value.String)
	if err != nil {
		return nil
	}
	return &parsed
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Truncate(time.Second).Format(time.RFC3339)
}

func nullableID(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func isUniqueError(err error) bool {
	return err != nil && (contains(err.Error(), "UNIQUE constraint failed") || contains(err.Error(), "constraint failed: UNIQUE"))
}

func contains(value, substring string) bool {
	if len(substring) == 0 || len(value) < len(substring) {
		return false
	}
	for i := 0; i+len(substring) <= len(value); i++ {
		if value[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}
