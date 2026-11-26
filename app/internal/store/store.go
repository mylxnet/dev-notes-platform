package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

const schemaVersion = 4

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

	// v1：建表。老库已是 v1 时跳过，新库依次走完全部版本。
	if current < 1 {
		for _, statement := range schemaStatements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply schema: %w", err)
			}
		}
	}

	// v2：项目增加说明与封面图，按列存在性判断，避免迁移中断后重跑报错。
	if current < 2 {
		columns := []struct {
			table string
			name  string
			spec  string
		}{
			{"project", "description", "TEXT NOT NULL DEFAULT ''"},
			{"project", "cover_image", "TEXT NOT NULL DEFAULT ''"},
		}
		for _, column := range columns {
			exists, err := columnExists(ctx, tx, column.table, column.name)
			if err != nil {
				return err
			}
			if exists {
				continue
			}
			statement := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", column.table, column.name, column.spec)
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("add column %s.%s: %w", column.table, column.name, err)
			}
		}
	}

	// v3：站点配置键值表，当前存备案号；新建库也走这里，与老库同一路径。
	if current < 3 {
		if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS settings (
			key        TEXT PRIMARY KEY,
			value      TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL
		)`); err != nil {
			return fmt.Errorf("create settings table: %w", err)
		}
	}

	// v4：推荐改为笔记自身属性，站点设置不再维护推荐位；迁移时把旧配置回填进来。
	if current < 4 {
		exists, err := columnExists(ctx, tx, "note", "is_featured")
		if err != nil {
			return err
		}
		if !exists {
			if _, err := tx.ExecContext(ctx,
				`ALTER TABLE note ADD COLUMN is_featured INTEGER NOT NULL DEFAULT 0`); err != nil {
				return fmt.Errorf("add column note.is_featured: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_note_featured
			ON note (is_featured, status, published_at DESC, id DESC)`); err != nil {
			return fmt.Errorf("create note featured index: %w", err)
		}
		if err := backfillFeaturedNotes(ctx, tx); err != nil {
			return err
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

// columnExists 用 PRAGMA table_info 判断列是否存在，让 ALTER TABLE 迁移可以安全重跑。
func columnExists(ctx context.Context, tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("read table info %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notNull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			return false, fmt.Errorf("scan table info %s: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

// backfillFeaturedNotes 把旧版存在 settings.featured_notes 里的推荐位迁到 note.is_featured。
// 旧值缺失或损坏都按「未配置」处理，不阻塞启动。
func backfillFeaturedNotes(ctx context.Context, tx *sql.Tx) error {
	var raw string
	err := tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = 'featured_notes'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read legacy featured notes: %w", err)
	}
	var slugs []string
	if err := json.Unmarshal([]byte(raw), &slugs); err != nil {
		return nil
	}
	for _, raw := range slugs {
		slug := strings.TrimSpace(raw)
		if slug == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE note SET is_featured = 1 WHERE slug = ?`, slug); err != nil {
			return fmt.Errorf("backfill featured note %s: %w", slug, err)
		}
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

// projectColumns 额外统计已发布笔记数，前台项目卡片据此展示，避免逐个请求详情。
const projectColumns = `id, name, slug, summary, tech_stack, repo_url, description, cover_image,
	sort_order, status, created_at, updated_at,
	(SELECT COUNT(*) FROM note WHERE note.project_id = project.id AND note.status = 'published')`

const noteColumns = `n.id, n.title, n.slug, n.summary, n.content, n.project_id, p.name, p.slug,
	n.is_featured, n.status, n.published_at, n.created_at, n.updated_at`

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
		&project.Description,
		&project.CoverImage,
		&project.SortOrder,
		&project.Status,
		&createdAt,
		&updatedAt,
		&project.NoteCount,
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
	var isFeatured int
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
		&isFeatured,
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
	note.IsFeatured = isFeatured == 1
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
