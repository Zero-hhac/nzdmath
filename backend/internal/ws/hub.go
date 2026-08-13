package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// 服务端 → 客户端的统一消息信封类型
const (
	TypeMessage      = "message"      // 新消息（含系统消息）
	TypePresence     = "presence"     // 在线人数变化
	TypeDelete       = "delete"       // 消息被撤回/删除
	TypePong         = "pong"         // 心跳回应
	TypeNotification = "notification" // 通知红点：有新通知（前端收到后刷新未读数）
)

const (
	WriteWait        = 10 * time.Second // 写超时
	ReadLimit        = 4096             // 单帧读取上限（bytes）
	ReadIdle         = 60 * time.Second // 读空闲超时：客户端每 30s 心跳，留一倍余量
	SendBuffer       = 64               // 每连接写缓冲（帧数）
	ClientPingEvery  = 30 * time.Second // 客户端心跳间隔
	ReconnectMaxWait = 30 * time.Second // 前端重连最长退避
)

// Envelope 是服务端 → 客户端推送的统一 JSON 信封
type Envelope struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Client 代表一个 WebSocket 连接
type Client struct {
	UserID uint
	Send   chan []byte
	conn   *websocket.Conn
	hub    *Hub
}

// Conn 暴露底层连接（供 handler 读写）
func (c *Client) Conn() *websocket.Conn { return c.conn }

// NewClient 创建并登记一个连接客户端
func NewClient(userID uint, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		UserID: userID,
		Send:   make(chan []byte, SendBuffer),
		conn:   conn,
		hub:    hub,
	}
}

// WritePump 是连接唯一的写 goroutine，负责将 Send 队列里的帧写回客户端。
// 通过单一写协程避免并发写 conn。
func (c *Client) WritePump() {
	defer c.conn.Close()
	for {
		payload, ok := <-c.Send
		_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
		if !ok {
			_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return
		}
	}
}

// Hub 维护所有在线连接，支持按用户分组与全局广播
type Hub struct {
	mu         sync.RWMutex
	clients    map[uint]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]map[*Client]struct{}),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
	}
}

// Run 是 Hub 的事件循环，在独立的 goroutine 中运行
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.clients[c.UserID] == nil {
				h.clients[c.UserID] = make(map[*Client]struct{})
			}
			h.clients[c.UserID][c] = struct{}{}
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if set, ok := h.clients[c.UserID]; ok {
				if _, exists := set[c]; exists {
					delete(set, c)
					close(c.Send)
				}
				if len(set) == 0 {
					delete(h.clients, c.UserID)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Register 将连接登记进 Hub（异步投递到事件循环）
func (h *Hub) Register(c *Client) { h.register <- c }

// Unregister 将连接移出 Hub
func (h *Hub) Unregister(c *Client) { h.unregister <- c }

// BroadcastMessage 推送一条新消息给所有在线连接
func (h *Hub) BroadcastMessage(data interface{}) {
	h.broadcast(TypeMessage, data)
}

// BroadcastPresence 推送在线人数变化
func (h *Hub) BroadcastPresence(onlineCount int64) {
	h.broadcast(TypePresence, map[string]int64{"online_count": onlineCount})
}

// BroadcastDelete 推送消息被撤回/删除
func (h *Hub) BroadcastDelete(ids []uint) {
	h.broadcast(TypeDelete, map[string][]uint{"ids": ids})
}

func (h *Hub) broadcast(t string, data interface{}) {
	payload, err := json.Marshal(Envelope{Type: t, Data: data})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, set := range h.clients {
		for c := range set {
			select {
			case c.Send <- payload:
			default:
				// 慢客户端：写缓冲已满，丢弃该帧（后续由心跳/超时断开）
			}
		}
	}
}

// SendToUser 给指定用户的所有连接推送消息（当前单实例下广播即可，保留以备扩展）
func (h *Hub) SendToUser(userID uint, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[userID] {
		select {
		case c.Send <- payload:
		default:
		}
	}
}
