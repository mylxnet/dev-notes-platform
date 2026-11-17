package model

import "time"

// Status 是内容的发布状态。
type Status string

const (
	StatusPublished Status = "published"
	StatusDraft     Status = "draft"
	StatusHidden    Status = "hidden"
)

// Valid 判断状态是否在允许范围内。
func (s Status) Valid() bool {
	switch s {
	case StatusPublished, StatusDraft, StatusHidden:
		return true
	default:
		return false
	}
}

// Project 对应 project 表。
type Project struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Summary   string    `json:"summary"`
	TechStack []string  `json:"tech_stack"`
	RepoURL   string    `json:"repo_url"`
	SortOrder int       `json:"sort_order"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProjectSummary 用于笔记详情里的关联项目摘要。
type ProjectSummary struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Note 对应 note 表。
type Note struct {
	ID          int64           `json:"id"`
	Title       string          `json:"title"`
	Slug        string          `json:"slug"`
	Summary     string          `json:"summary"`
	Content     string          `json:"content,omitempty"`
	ProjectID   *int64          `json:"project_id,omitempty"`
	Project     *ProjectSummary `json:"project,omitempty"`
	Status      Status          `json:"status"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Admin 对应 admin 表。
type Admin struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	UpdatedAt    time.Time `json:"updated_at"`
}
