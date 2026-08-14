package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// fallbackWindow / fallbackScale：Redis 不可用时进程内存兜底限流参数。
// 兜底阈值放宽容限（limit * 4），只做粗粒度防护，避免误伤正常流量。
const (
	fallbackWindow = time.Minute
	fallbackScale  = 4
)

type fallbackCounter struct {
	count       int64
	windowStart time.Time
}

var (
	fallbackMu       sync.Mutex
	fallbackCounters = make(map[string]*fallbackCounter)
)

// RateLimitMiddleware implements a fixed-window rate limit backed by Redis so
// the limit stays consistent across multiple backend instances.
// #17：Redis 故障时降级为进程内存计数器做粗粒度兜底限流，不再 fail-open 完全放行。
func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			// 无 Redis 配置（单测等场景）：直接放行
			c.Next()
			return
		}

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		key := fmt.Sprintf("ratelimit:%s:%s", route, c.ClientIP())
		ctx := c.Request.Context()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis 不可用：降级到进程内存兜底限流（fail-closed 而非 fail-open）
			if fallbackAllow(key, limit) {
				c.Next()
			} else {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"code": 429,
					"msg":  "请求过于频繁，请稍后再试",
				})
			}
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// fallbackAllow 进程内存粗粒度限流：滑动窗口内超过 limit*fallbackScale 即拒绝。
// 定期清理过期计数，防止内存无限增长。
func fallbackAllow(key string, limit int) bool {
	now := time.Now()
	fallbackMu.Lock()
	defer fallbackMu.Unlock()

	// 惰性清理超过一个窗口的过期计数（每 256 次触发一次全量清理）
	if now.UnixNano()%256 == 0 {
		for k, v := range fallbackCounters {
			if now.Sub(v.windowStart) > fallbackWindow*2 {
				delete(fallbackCounters, k)
			}
		}
	}

	entry, ok := fallbackCounters[key]
	if !ok || now.Sub(entry.windowStart) >= fallbackWindow {
		fallbackCounters[key] = &fallbackCounter{count: 1, windowStart: now}
		return true
	}
	entry.count++
	return entry.count <= int64(limit*fallbackScale)
}
