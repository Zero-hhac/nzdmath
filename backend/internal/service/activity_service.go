package service

import (
	"context"
	"math-top/internal/middleware"
	"math-top/internal/utils"
	"time"

	"github.com/redis/go-redis/v9"
)

type ActivityService struct {
	rdb *redis.Client
}

func NewActivityService(rdb *redis.Client) *ActivityService {
	return &ActivityService{rdb: rdb}
}

// Track 记录一次页面访问：PV 计数、UV（按 IP 去重）、DAU（按登录用户去重）。
// 日 key 的 TTL 设为 8 天：给 SyncDailyMetrics 的每小时归档留足缓冲，
// 防止管理员几天不登录后台导致 48 小时过期、数据永久丢失。
func (s *ActivityService) Track(ctx context.Context, clientIP, tokenString string) {
	today := time.Now().Format("2006-01-02")
	const dayKeyTTL = 8 * 24 * time.Hour

	// 1. 统计 PV (Page View)
	pvKey := "dau:pv:" + today
	s.rdb.Incr(ctx, pvKey)
	s.rdb.Expire(ctx, pvKey, dayKeyTTL)
	s.rdb.Incr(ctx, "dau:pv:all")

	// 2. 统计 UV (IP 独立访客)
	if clientIP != "" {
		uvKey := "dau:ip:" + today
		s.rdb.PFAdd(ctx, uvKey, clientIP)
		s.rdb.Expire(ctx, uvKey, dayKeyTTL)
		s.rdb.PFAdd(ctx, "dau:ip:all", clientIP)
	}

	// 3. 统计 DAU (会员独立活跃数)
	if userID := s.resolveUserID(ctx, tokenString); userID > 0 {
		dauKey := "dau:user:" + today
		s.rdb.PFAdd(ctx, dauKey, userID)
		s.rdb.Expire(ctx, dauKey, dayKeyTTL)
		s.rdb.PFAdd(ctx, "dau:user:all", userID)
	}
}

func (s *ActivityService) resolveUserID(ctx context.Context, tokenString string) uint {
	if tokenString == "" {
		return 0
	}
	redisKey := middleware.UserTokenPrefix + tokenString
	if s.rdb.Exists(ctx, redisKey).Val() == 0 {
		return 0
	}
	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		return 0
	}
	return claims.UserID
}
