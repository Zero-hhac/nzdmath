package middleware

import (
	"math-top/internal/response"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Fail(c, 401, "未登录")
			c.Abort()
			return
		}
		roleStr, ok := role.(string)
		if !ok || (roleStr != "admin" && roleStr != "super_admin") {
			response.Fail(c, 403, "无权访问，需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
