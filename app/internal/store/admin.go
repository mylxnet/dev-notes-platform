package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
