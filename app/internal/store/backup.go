package store

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
)

// mediaRefPattern 匹配正文/封面里引用的运行时上传图片。
// 上传文件名固定为 32 位十六进制加扩展名，见 api/media.go。
var mediaRefPattern = regexp.MustCompile(`/media/uploads/([0-9a-f]{32}\.(?:png|jpg|jpeg|webp|gif))`)

// CurrentSchemaVersion 返回当前程序支持的数据库结构版本，备份预检用它判断兼容性。
func CurrentSchemaVersion() int { return schemaVersion }

// SnapshotTo 用 VACUUM INTO 导出一致性快照。
//
// 不能直接复制 notes.db：WAL 模式下未 checkpoint 的事务还留在 -wal 文件里，
// 只拷主库会丢掉最近的写入。VACUUM INTO 由 SQLite 自己生成一份干净的库文件。
func (s *Store) SnapshotTo(ctx context.Context, dest string) error {
	if dest == "" {
		return errors.New("snapshot path is empty")
	}
	if err := os.Remove(dest); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("clean snapshot target: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, "VACUUM INTO ?", dest); err != nil {
		return fmt.Errorf("vacuum into snapshot: %w", err)
	}
	return nil
}

// SchemaVersionOf 返回数据库结构版本，写进备份清单用。
func (s *Store) SchemaVersionOf(ctx context.Context) (int, error) {
	var version int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return 0, fmt.Errorf("read user_version: %w", err)
	}
	return version, nil
}

// ContentCounts 统计笔记与项目数量，写进备份清单用。
func (s *Store) ContentCounts(ctx context.Context) (notes int, projects int, err error) {
	if err := s.db.QueryRowContext(ctx,
		`SELECT (SELECT COUNT(*) FROM note), (SELECT COUNT(*) FROM project)`).Scan(&notes, &projects); err != nil {
		return 0, 0, fmt.Errorf("count content: %w", err)
	}
	return notes, projects, nil
}

// MediaReferences 收集内容里引用到的上传图片文件名（去重、有序）。
// 恢复后用它在磁盘上逐个核对，回答「内容是否都能正常显示」。
func (s *Store) MediaReferences(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT content FROM note
		UNION ALL SELECT description FROM project
		UNION ALL SELECT cover_image FROM project`)
	if err != nil {
		return nil, fmt.Errorf("scan media references: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, fmt.Errorf("scan media reference row: %w", err)
		}
		for _, match := range mediaRefPattern.FindAllStringSubmatch(text, -1) {
			seen[match[1]] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate media references: %w", err)
	}

	refs := make([]string, 0, len(seen))
	for name := range seen {
		refs = append(refs, name)
	}
	sort.Strings(refs)
	return refs, nil
}
