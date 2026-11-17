package api

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// 单张图片最大 2 MB；multipart 额外开销由路由的 bodyLimit 兜底。
	maxUploadBytes = 2 << 20
)

var (
	allowedUploadExt = map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".webp": true,
	}
	allowedUploadMIME = map[string]bool{
		"image/png": true, "image/jpeg": true, "image/webp": true,
	}
	// 上传后的文件名只由服务端生成，读取时同样严格校验，杜绝目录穿越。
	uploadedNamePattern = regexp.MustCompile(`^[a-f0-9]{32}\.(png|jpg|jpeg|webp)$`)
)

// adminUploadMedia 保存后台上传的封面图，返回可直接填进封面字段的站内路径。
func (s *Server) adminUploadMedia(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请选择要上传的图片")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedUploadExt[ext] {
		respondError(c, http.StatusUnprocessableEntity, "UNSUPPORTED_IMAGE", "只支持 PNG、JPG、WEBP 图片")
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "读取图片失败")
		return
	}
	if len(data) > maxUploadBytes {
		respondError(c, http.StatusUnprocessableEntity, "IMAGE_TOO_LARGE", "图片不能超过 2 MB")
		return
	}
	if len(data) == 0 || !allowedUploadMIME[http.DetectContentType(data)] {
		respondError(c, http.StatusUnprocessableEntity, "UNSUPPORTED_IMAGE", "文件内容不是有效的 PNG、JPG、WEBP 图片")
		return
	}

	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "生成图片名失败")
		return
	}
	name := hex.EncodeToString(raw[:]) + ext

	if err := os.MkdirAll(s.mediaDir, 0o755); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建图片目录失败")
		return
	}
	tmpPath := filepath.Join(s.mediaDir, name+".part")
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "保存图片失败")
		return
	}
	if err := os.Rename(tmpPath, filepath.Join(s.mediaDir, name)); err != nil {
		_ = os.Remove(tmpPath)
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "保存图片失败")
		return
	}

	respondOK(c, gin.H{"path": "/media/uploads/" + name})
}

// serveUploadedMedia 从数据目录下发运行时上传的图片（构建产物里的 /media/ 由静态回退处理）。
func (s *Server) serveUploadedMedia(c *gin.Context) {
	name := c.Param("name")
	if !uploadedNamePattern.MatchString(name) {
		c.Status(http.StatusNotFound)
		return
	}
	filePath := filepath.Join(s.mediaDir, name)
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}
	// 文件名不可变，可以长缓存；安全响应头由全局中间件设置。
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(filePath)
}
