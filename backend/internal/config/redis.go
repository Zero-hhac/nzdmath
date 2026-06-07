package config

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	cfg := GlobalConfig
	if cfg == nil {
		LoadConfig()
		cfg = GlobalConfig
	}
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		slog.Error("Redis 连接失败", "addr", addr, "err", err)
		panic(err)
	}
	slog.Info("Redis 连接成功", "addr", addr)
	return rdb
}