//go:build embedui

package webui

import (
	"embed"
	"io/fs"
)

// 构建脚本（scripts/build.sh）会先把 web/ 产物输出到 dist/，再用 -tags embedui 编译。
//
//go:embed all:dist
var embedded embed.FS

func embeddedFS() (fs.FS, bool) {
	return embedded, true
}
