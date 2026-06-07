package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) Get(ctx context.Context, key string, dest any) bool {
	if c == nil || c.rdb == nil {
		return false
	}
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}
	if err := json.Unmarshal(data, dest); err != nil {
		slog.Warn("缓存反序列化失败", "key", key, "err", err)
		return false
	}
	return true
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	if c == nil || c.rdb == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		slog.Warn("缓存序列化失败", "key", key, "err", err)
		return
	}
	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		slog.Warn("缓存写入失败", "key", key, "err", err)
	}
}

func (c *Cache) Del(ctx context.Context, keys ...string) {
	if c == nil || c.rdb == nil || len(keys) == 0 {
		return
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		slog.Warn("缓存删除失败", "keys", keys, "err", err)
	}
}

func (c *Cache) Incr(ctx context.Context, key string) int64 {
	if c == nil || c.rdb == nil {
		return 0
	}
	n, _ := c.rdb.Incr(ctx, key).Result()
	return n
}

func (c *Cache) GetInt(ctx context.Context, key string) int {
	if c == nil || c.rdb == nil {
		return 0
	}
	v, err := c.rdb.Get(ctx, key).Int()
	if err != nil {
		return 0
	}
	return v
}