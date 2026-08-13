package handler

import (
	"encoding/json"
	"math-top/internal/middleware"
	"math-top/internal/response"
	"math-top/internal/ws"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// NotifyWSHandler 通知实时推送通道：与聊天室共用同一个 Hub，
// 但只接收通知事件、不触发聊天室 Join/Presence，避免污染聊天室在线状态。
// 鉴权与聊天 WS 完全一致：子协议携带 token，同源校验。
type NotifyWSHandler struct {
	hub *ws.Hub
	rdb *redis.Client
}

func NewNotifyWSHandler(hub *ws.Hub, rdb *redis.Client) *NotifyWSHandler {
	return &NotifyWSHandler{hub: hub, rdb: rdb}
}

// Handle 升级 WebSocket：鉴权 → 注册进 Hub（不 Join 聊天室）→ 读循环保活
func (h *NotifyWSHandler) Handle(c *gin.Context) {
	token := ""
	for _, p := range websocket.Subprotocols(c.Request) {
		if strings.HasPrefix(p, "bearer.") {
			token = strings.TrimPrefix(p, "bearer.")
			break
		}
	}
	if token == "" {
		response.Fail(c, http.StatusUnauthorized, "缺少 token")
		return
	}
	claims, err := middleware.AuthenticateToken(h.rdb, token, middleware.UserTokenPrefix)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			return strings.EqualFold(u.Host, r.Host)
		},
		Subprotocols: websocket.Subprotocols(c.Request),
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(claims.UserID, conn, h.hub)
	h.hub.Register(client)
	go client.WritePump()

	h.readLoop(client)
}

// readLoop 仅保活：处理客户端 ping；断开时从 Hub 移除
func (h *NotifyWSHandler) readLoop(client *ws.Client) {
	defer func() {
		h.hub.Unregister(client)
		_ = client.Conn().Close()
	}()

	client.Conn().SetReadLimit(ws.ReadLimit)
	for {
		_ = client.Conn().SetReadDeadline(time.Now().Add(ws.ReadIdle))
		_, data, err := client.Conn().ReadMessage()
		if err != nil {
			return
		}
		var req struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(data, &req) != nil {
			continue
		}
		if req.Type == "ping" {
			payload, _ := json.Marshal(ws.Envelope{Type: ws.TypePong, Data: map[string]int64{"ts": time.Now().UnixMilli()}})
			select {
			case client.Send <- payload:
			default:
			}
		}
	}
}
