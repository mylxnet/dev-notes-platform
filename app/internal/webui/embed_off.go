//go:build !embedui

package webui

import "io/fs"

// 默认构建不嵌入前端资源：本地开发由 Vite 提供页面，后端只提供 API。
func embeddedFS() (fs.FS, bool) {
	return nil, false
}
