package middleware

import (
	"context"
	"errors"
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

// AuthenticateToken 校验 token 是否有效：Redis 中存在（未登出/未过期）且 JWT 签名正确。
// HTTP 中间件与 WebSocket 握手共用，避免两套校验逻辑漂移。
func AuthenticateToken(rdb *redis.Client, tokenString, prefix string) (*utils.MyClaims, error) {
	if rdb == nil || tokenString == "" {
		return nil, errors.New("token 无效或者已过期")
	}
	ctx := context.Background()
	redisKey := prefix + tokenString
	if rdb.Exists(ctx, redisKey).Val() == 0 {
		return nil, errors.New("token 无效或者已过期")
	}
	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		return nil, errors.New("token 无效或者已过期")
	}
	return claims, nil
}

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

		claims, err := AuthenticateToken(rdb, tokenString, prefix)
		if err != nil {
			response.Fail(c, 401, err.Error())
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
