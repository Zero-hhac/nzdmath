package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func newRateLimitServer(limit int, window time.Duration) (*gin.Engine, *redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(rdb, limit, window))
	r.GET("/limited", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r, rdb, mr
}

func doRequest(r *gin.Engine) int {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRateLimitBlocksAfterLimit(t *testing.T) {
	r, _, mr := newRateLimitServer(3, time.Minute)
	defer mr.Close()

	for i := 0; i < 3; i++ {
		if code := doRequest(r); code != http.StatusOK {
			t.Fatalf("第 %d 次请求状态 = %d, want 200", i+1, code)
		}
	}
	if code := doRequest(r); code != http.StatusTooManyRequests {
		t.Fatalf("超限请求状态 = %d, want 429", code)
	}
}

func TestRateLimitAllowsWhenRedisDown(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	addr := mr.Addr()
	mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: addr})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(rdb, 1, time.Minute))
	r.GET("/limited", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Redis 不可用时不应阻塞请求（失败放行）
	for i := 0; i < 3; i++ {
		if code := doRequest(r); code != http.StatusOK {
			t.Fatalf("Redis 不可用时第 %d 次请求状态 = %d, want 200", i+1, code)
		}
	}
}

func TestRateLimitNilRedisPassThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(nil, 1, time.Minute))
	r.GET("/limited", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	if code := doRequest(r); code != http.StatusOK {
		t.Fatalf("nil Redis 时请求状态 = %d, want 200", code)
	}
}
