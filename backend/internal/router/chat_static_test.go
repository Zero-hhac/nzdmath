package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"math-top/internal/config"
	"math-top/internal/middleware"
	"math-top/internal/utils"
)

// TestChatUploadsRequireAuth 验证上传目录白名单挂载：
// - 公共目录（avatars 等）无需鉴权即可访问；
// - /uploads/chat 必须带有效用户 token（Redis 中存在），否则 401；
// - 未列入白名单的目录 404。
// 注意：gin 的路由树不允许 /uploads/chat/*filepath 与 /uploads/*filepath 同时注册
// （会 panic），因此生产代码采用逐目录白名单挂载而非整根挂载 + 影子路由。
func TestChatUploadsRequireAuth(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dir := t.TempDir()
	writeFile := func(rel string, content string) {
		t.Helper()
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeFile("avatars/a.png", "avatar")
	writeFile("chat/secret.txt", "secret")

	// 生成一个在 Redis 中登记过的有效 token
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{Secret: "unit-test-secret-01234567890123456789", ExpireHours: 24},
	}
	token, err := utils.GenerateToken(1, "alice", "member")
	if err != nil {
		t.Fatal(err)
	}
	mr.Set(middleware.UserTokenPrefix+token, "1")

	gin.SetMode(gin.TestMode)
	r := gin.New()

	publicDirs := []string{"avatars", "covers", "h5_unified_light", "resources"}
	for _, publicDir := range publicDirs {
		r.Static("/uploads/"+publicDir, filepath.Join(dir, publicDir))
	}
	chat := r.Group("/uploads/chat")
	chat.Use(middleware.JWTAuthMiddleware(rdb))
	chat.Static("", filepath.Join(dir, "chat"))

	t.Run("public dir needs no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/uploads/avatars/a.png", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if w.Body.String() != "avatar" {
			t.Fatalf("unexpected body: %q", w.Body.String())
		}
	})

	t.Run("unlisted dir is 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/uploads/unknown/x.png", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("chat uploads require auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/uploads/chat/secret.txt", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 without token, got %d", w.Code)
		}
	})

	t.Run("chat uploads allowed with valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/uploads/chat/secret.txt", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 with valid token, got %d", w.Code)
		}
		if w.Body.String() != "secret" {
			t.Fatalf("unexpected body: %q", w.Body.String())
		}
	})
}
