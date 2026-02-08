package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var webFS embed.FS

// GetWebFS 获取嵌入的 Web 文件系统
// 返回 web 子目录（包含 dist）
func GetWebFS() (fs.FS, error) {
	return fs.Sub(webFS, "web")
}
