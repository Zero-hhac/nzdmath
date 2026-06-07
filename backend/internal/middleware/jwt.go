package middleware

import (
	"context"
	"math-top/internal/response"
	"math-top/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	UserTokenPrefix  = "token:"
	AdminTokenPrefix = "admin_token:"
)

func JWTAuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return AuthMiddlewareWithPrefix(rdb, UserTokenPrefix)
}

func AdminJWTAuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return AuthMiddlewareWithPrefix(rdb, AdminTokenPrefix)
}

func AuthMiddlewareWithPrefix(rdb *redis.Client, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		AutoHeader := c.GetHeader("Authorization")
		if AutoHeader == "" {
			response.Fail(c, 401, "请求头中缺少 Authorization")
			c.Abort()
			return
		}
		parts := strings.Split(AutoHeader, " ")
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Fail(c, 401, "请求头中 Authorization 格式有误")
			c.Abort()
			return
		}
		tokenString := parts[1]

		ctx := context.Background()
		redisKey := prefix + tokenString
		if rdb.Exists(ctx, redisKey).Val() == 0 {
			response.Fail(c, 401, "token 无效或者已过期")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			response.Fail(c, 401, "token 无效或者已过期")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
