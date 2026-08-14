package service

import (
	"context"
	"strconv"
	"time"

	"math-top/internal/middleware"
	"math-top/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// tokenIndexUserKey / tokenIndexAdminKey：用户 → 已签发 token 集合的 Redis 索引，
// 与白名单 token key 同 TTL，用于改密/重置/封禁/删除后统一吊销该用户全部会话。
const (
	tokenIndexUserKey  = "utok:"
	tokenIndexAdminKey = "atok:"
)

func tokenIndexKey(prefix string, userID uint) string {
	if prefix == middleware.AdminTokenPrefix {
		return tokenIndexAdminKey + strconv.FormatUint(uint64(userID), 10)
	}
	return tokenIndexUserKey + strconv.FormatUint(uint64(userID), 10)
}

// IndexUserToken 登录签发 token 时登记进用户索引（与白名单 key 同 TTL）。
func IndexUserToken(rdb *redis.Client, userID uint, token, prefix string, ttl time.Duration) {
	if rdb == nil || token == "" {
		return
	}
	ctx := context.Background()
	key := tokenIndexKey(prefix, userID)
	if err := rdb.SAdd(ctx, key, token).Err(); err == nil {
		rdb.Expire(ctx, key, ttl)
	}
}

// UnindexUserToken 登出时从用户索引移除该 token。
func UnindexUserToken(rdb *redis.Client, userID uint, token, prefix string) {
	if rdb == nil || token == "" {
		return
	}
	rdb.SRem(context.Background(), tokenIndexKey(prefix, userID), token)
}

// RevokeUserTokens 吊销指定用户 ID 的全部会话：取出索引中所有 token，
// 逐个删除白名单 key，再删除索引本身。
func RevokeUserTokens(rdb *redis.Client, userID uint) {
	if rdb == nil {
		return
	}
	ctx := context.Background()
	for _, idx := range []string{tokenIndexUserKey, tokenIndexAdminKey} {
		key := idx + strconv.FormatUint(uint64(userID), 10)
		tokens, err := rdb.SMembers(ctx, key).Result()
		if err != nil {
			rdb.Del(ctx, key)
			continue
		}
		prefix := middleware.UserTokenPrefix
		if idx == tokenIndexAdminKey {
			prefix = middleware.AdminTokenPrefix
		}
		for _, t := range tokens {
			rdb.Del(ctx, prefix+t)
		}
		rdb.Del(ctx, key)
	}
}

// RevokeTokensForUserID 吊销会员表用户及其同名管理员账号（admins 表）的全部会话。
// 管理员登录签发的 admin_token 携带 admins 表 ID，需要按用户名联动吊销。
func RevokeTokensForUserID(rdb *redis.Client, db *gorm.DB, userID uint) {
	if rdb == nil || db == nil {
		return
	}
	var user model.User
	if err := db.First(&user, userID).Error; err == nil {
		var admin model.Admin
		if err := db.Where("username = ?", user.Username).First(&admin).Error; err == nil {
			RevokeUserTokens(rdb, admin.ID)
		}
	}
	RevokeUserTokens(rdb, userID)
}
