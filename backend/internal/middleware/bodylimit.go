package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 请求体大小上限：JSON 接口 1MB，multipart 上传 64MB（资源文件最大 50MB，留余量）。
// 防止无上限的请求体造成内存 DoS（此前实测 5MB JSON 可直通解析）。
const (
	MaxJSONBody      = 1 << 20 // 1MB
	MaxMultipartBody = 64 << 20
)

// BodyLimitMiddleware 按 Content-Type 分档限制请求体大小，超限时读取会报错，
// 由各 handler 的 ShouldBindJSON / FormFile 统一转为 400。
func BodyLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := int64(MaxJSONBody)
		if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
			limit = MaxMultipartBody
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
