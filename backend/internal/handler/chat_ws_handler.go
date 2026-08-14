package handler

import (
	"encoding/json"
	"math-top/internal/middleware"
	"math-top/internal/response"
	"math-top/internal/service"
	"math-top/internal/ws"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// ChatWSHandler 处理聊天室 WebSocket 连接：握手鉴权 + 加入/离开 + 读写循环
type ChatWSHandler struct {
	svc *service.ChatService
	hub *ws.Hub
	rdb *redis.Client
}

func NewChatWSHandler(svc *service.ChatService, hub *ws.Hub, rdb *redis.Client) *ChatWSHandler {
	return &ChatWSHandler{svc: svc, hub: hub, rdb: rdb}
}

// Handle 升级 WebSocket 连接。
// 鉴权说明：浏览器 WebSocket 无法自定义 Authorization header，
// 采用子协议方案 —— 前端 new WebSocket(url, ['bearer.<token>'])，
// token 经 Sec-WebSocket-Protocol 传递，不落入 URL / access log。
func (h *ChatWSHandler) Handle(c *gin.Context) {
	token := ""
	for _, p := range websocket.Subprotocols(c.Request) {
		if strings.HasPrefix(p, "bearer.") {
			token = strings.TrimPrefix(p, "bearer.")
			break
		}
	}
	if token == "" {
		response.Fail(c, 401, "缺少 token")
		return
	}
	claims, err := middleware.AuthenticateToken(h.rdb, token, middleware.UserTokenPrefix)
	if err != nil {
		response.Fail(c, 401, err.Error())
		return
	}

	upgrader := websocket.Upgrader{
		// 同源校验：浏览器请求必须与 Host 同源，防止跨站 WebSocket 连接；
		// 无 Origin 头的非浏览器客户端（如 Python 脚本）放行。
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
		// 必须回显浏览器请求的子协议（bearer.<token>）：gorilla 仅在配置
		// Subprotocols 时才在 101 响应中带 Sec-WebSocket-Protocol 头，
		// 浏览器 WebSocket 要求回显，否则判定握手失败；python 客户端不强制。
		Subprotocols: websocket.Subprotocols(c.Request),
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(claims.UserID, conn, h.hub)
	h.hub.Register(client)
	go client.WritePump()

	// 先注册再 Join：确保「xxx加入聊天室」系统消息与在线数广播能推给本连接
	count, err := h.svc.Join(claims.UserID)
	if err != nil {
		h.hub.Unregister(client)
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, err.Error()),
			time.Now().Add(5*time.Second))
		_ = conn.Close()
		return
	}
	h.hub.BroadcastPresence(count)

	h.readLoop(client)
}

// #27 WebSocket 发消息限频参数：10 秒滑动窗口内至多 10 条，
// 超限先推送警告，连续超限则断开连接。
const (
	wsRateWindow   = 10 * time.Second
	wsRateMax      = 10
	wsRateWarnOnce = 3 // 连续超限多少次后断开
)

// readLoop 是连接唯一的读循环：处理客户端 ping / 发消息，断开时清理并广播在线数
func (h *ChatWSHandler) readLoop(client *ws.Client) {
	defer func() {
		h.hub.Unregister(client)
		_ = h.svc.Leave(client.UserID)
		if count, err := h.svc.OnlineCount(); err == nil {
			h.hub.BroadcastPresence(count)
		}
		_ = client.Conn().Close()
	}()

	client.Conn().SetReadLimit(ws.ReadLimit)

	type wsRequest struct {
		Type string `json:"type"`
		Data struct {
			Content string `json:"content"`
			TS      int64  `json:"ts"`
		} `json:"data"`
	}

	// #27 滑动窗口限频：记录每条消息的发送时间戳
	var sendTimes []time.Time
	overLimitStreak := 0

	rateAllowed := func() bool {
		now := time.Now()
		cutoff := now.Add(-wsRateWindow)
		kept := sendTimes[:0]
		for _, t := range sendTimes {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		sendTimes = kept
		if len(sendTimes) >= wsRateMax {
			return false
		}
		sendTimes = append(sendTimes, now)
		return true
	}

	pushWarning := func() {
		payload, _ := json.Marshal(ws.Envelope{
			Type: "rate_warn",
			Data: map[string]string{"message": "发送消息过于频繁，请稍后再试"},
		})
		select {
		case client.Send <- payload:
		default:
		}
	}

	for {
		_ = client.Conn().SetReadDeadline(time.Now().Add(ws.ReadIdle))
		_, data, err := client.Conn().ReadMessage()
		if err != nil {
			return
		}
		var req wsRequest
		if json.Unmarshal(data, &req) != nil {
			continue
		}
		switch req.Type {
		case "ping":
			payload, _ := json.Marshal(ws.Envelope{Type: ws.TypePong, Data: map[string]int64{"ts": req.Data.TS}})
			select {
			case client.Send <- payload:
			default:
			}
		case "message":
			content := strings.TrimSpace(req.Data.Content)
			if content == "" {
				continue
			}
			// #27：滑动窗口限频，超限先推送警告，持续超限断开连接
			if !rateAllowed() {
				overLimitStreak++
				pushWarning()
				if overLimitStreak >= wsRateWarnOnce {
					_ = client.Conn().WriteControl(websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "发送消息过于频繁"),
						time.Now().Add(5*time.Second))
					return
				}
				continue
			}
			overLimitStreak = 0
			// 发送成功由 service 内部广播给所有人（含发送者，前端按 id 去重）
			_, _ = h.svc.SendText(client.UserID, content)
		}
	}
}
