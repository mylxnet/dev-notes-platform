// Package webui 提供前端构建产物的静态资源入口。
//
// 正式构建使用 `-tags embedui` 把 dist 目录嵌入二进制（见 embed_on.go）；
// 默认构建不嵌入，只提供 API，前端由 Vite 开发服务器提供。
package webui

import "io/fs"

// FS 返回以前端产物根目录（dist）为根的文件系统；未嵌入资源时返回 nil。
func FS() fs.FS {
	base, ok := embeddedFS()
	if !ok || base == nil {
		return nil
	}
	sub, err := fs.Sub(base, "dist")
	if err != nil {
		return nil
	}
	return sub
}
