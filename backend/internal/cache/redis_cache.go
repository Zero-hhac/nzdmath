package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
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

func (c *Cache) SetNX(ctx context.Context, key string, value any, ttl time.Duration) bool {
	if c == nil || c.rdb == nil {
		return true
	}
	ok, err := c.rdb.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return true
	}
	return ok
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

// SetInt stores a primitive int with a TTL.
func (c *Cache) SetInt(ctx context.Context, key string, val int64, ttl time.Duration) {
	if c == nil || c.rdb == nil {
		return
	}
	if err := c.rdb.Set(ctx, key, val, ttl).Err(); err != nil {
		slog.Warn("缓存写入失败", "key", key, "err", err)
	}
}

// ZAdd adds a member with the given score to a sorted set.
func (c *Cache) ZAdd(ctx context.Context, key, member string, score float64) bool {
	if c == nil || c.rdb == nil {
		return false
	}
	if err := c.rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err(); err != nil {
		slog.Warn("ZAdd 失败", "key", key, "err", err)
		return false
	}
	return true
}

// ZRangeByScore returns members with score in (min, max], ordered by score asc.
func (c *Cache) ZRangeByScore(ctx context.Context, key string, min, max float64) []string {
	if c == nil || c.rdb == nil {
		return nil
	}
	val, err := c.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: strconv.FormatFloat(min, 'f', -1, 64), Max: strconv.FormatFloat(max, 'f', -1, 64)}).Result()
	if err != nil {
		return nil
	}
	return val
}

// ZRangeByRank returns members ordered by score, including both rank bounds.
func (c *Cache) ZRangeByRank(ctx context.Context, key string, start, stop int64) ([]string, bool) {
	if c == nil || c.rdb == nil {
		return nil, false
	}
	val, err := c.rdb.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

// ZScore distinguishes a missing member from an unavailable Redis connection.
func (c *Cache) ZScore(ctx context.Context, key, member string) (float64, bool, bool) {
	if c == nil || c.rdb == nil {
		return 0, false, false
	}
	val, err := c.rdb.ZScore(ctx, key, member).Result()
	if err == redis.Nil {
		return 0, false, true
	}
	if err != nil {
		return 0, false, false
	}
	return val, true, true
}

// ZCount returns the number of members in the inclusive score range.
func (c *Cache) ZCount(ctx context.Context, key string, min, max float64) (int64, bool) {
	if c == nil || c.rdb == nil {
		return 0, false
	}
	val, err := c.rdb.ZCount(ctx, key, strconv.FormatFloat(min, 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64)).Result()
	if err != nil {
		return 0, false
	}
	return val, true
}

// ZRem removes members from a sorted set.
func (c *Cache) ZRem(ctx context.Context, key string, members ...string) {
	if c == nil || c.rdb == nil || len(members) == 0 {
		return
	}
	if err := c.rdb.ZRem(ctx, key, members).Err(); err != nil {
		slog.Warn("ZRem 失败", "key", key, "err", err)
	}
}

// ZRemRangeByRank removes members by rank range.
func (c *Cache) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) {
	if c == nil || c.rdb == nil {
		return
	}
	c.rdb.ZRemRangeByRank(ctx, key, start, stop)
}

// ZRemRangeByScore removes members whose score is in [min, max].
func (c *Cache) ZRemRangeByScore(ctx context.Context, key string, min, max float64) {
	if c == nil || c.rdb == nil {
		return
	}
	c.rdb.ZRemRangeByScore(ctx, key, strconv.FormatFloat(min, 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
}
