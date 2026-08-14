package router

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// staticWithDownloadOnlyForHTML 包装公开静态目录：
// 对 HTML / HTM / SVG 后缀一律强制附件下载并改写 Content-Type，
// 该目录任何文件不再以网页形式内联执行（防同源执行面，见 #9）。
// 其余文件行为与原静态挂载完全一致。
func staticWithDownloadOnlyForHTML(urlPrefix, dir string) gin.HandlerFunc {
	fs := http.FileServer(http.Dir(dir))
	return func(c *gin.Context) {
		reqPath := c.Param("filepath")
		ext := strings.ToLower(filepath.Ext(reqPath))
		if ext == ".html" || ext == ".htm" || ext == ".svg" {
			// 手工落盘：先做目录穿越防护，再以附件形式下发
			full := filepath.Clean(filepath.Join(dir, filepath.FromSlash(reqPath)))
			baseAbs, err := filepath.Abs(dir)
			if err != nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			if !strings.HasPrefix(filepath.Clean(full), baseAbs+string(filepath.Separator)) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename="+filepath.Base(reqPath))
			c.File(full)
			return
		}
		http.StripPrefix(urlPrefix, fs).ServeHTTP(c.Writer, c.Request)
	}
}
