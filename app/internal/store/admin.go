package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dev-notes-platform/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// AdminCount 返回管理员数量。
func (s *Store) AdminCount(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return count, nil
}

// GetAdminByUsername 按账号查询管理员。
func (s *Store) GetAdminByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	var updatedAt string
	err := s.db.QueryRowContext(ctx, `SELECT id, username, password_hash, updated_at
		FROM admin WHERE username = ?`, username).
		Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get admin: %w", err)
	}
	admin.UpdatedAt = parseTime(updatedAt)
	return &admin, nil
}

// GetAdminByID 按 ID 查询管理员，供会话校验使用。
func (s *Store) GetAdminByID(ctx context.Context, id int64) (*model.Admin, error) {
	var admin model.Admin
	var updatedAt string
	err := s.db.QueryRowContext(ctx, `SELECT id, username, password_hash, updated_at
		FROM admin WHERE id = ?`, id).
		Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get admin by id: %w", err)
	}
	admin.UpdatedAt = parseTime(updatedAt)
	return &admin, nil
}

// CreateAdmin 创建管理员并写入 bcrypt 哈希。
func (s *Store) CreateAdmin(ctx context.Context, username, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}
	now := nowText()
	result, err := s.db.ExecContext(ctx, `INSERT INTO admin (username, password_hash, updated_at)
		VALUES (?, ?, ?)`, username, string(hash), now)
	if err != nil {
		if isUniqueError(err) {
			return 0, ErrConflict
		}
		return 0, fmt.Errorf("create admin: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("admin id: %w", err)
	}
	return id, nil
}

// SeedAdmin 在 admin 表为空时创建初始管理员。
func (s *Store) SeedAdmin(ctx context.Context, username, password string) error {
	count, err := s.AdminCount(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err = s.CreateAdmin(ctx, username, password)
	return err
}

// UpdateAdminPassword 更新密码哈希。
// updated_at 单调递增，保证改密后旧会话（携带旧时间戳）立即失效。
func (s *Store) UpdateAdminPassword(ctx context.Context, id int64, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	var current string
	err = s.db.QueryRowContext(ctx, `SELECT updated_at FROM admin WHERE id = ?`, id).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("read admin updated_at: %w", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	if prev := parseTime(current); !prev.IsZero() && !now.After(prev) {
		now = prev.Add(time.Second)
	}

	if _, err := s.db.ExecContext(ctx,
		`UPDATE admin SET password_hash = ?, updated_at = ? WHERE id = ?`,
		string(hash), now.Format(time.RFC3339), id); err != nil {
		return fmt.Errorf("update admin password: %w", err)
	}
	return nil
}
