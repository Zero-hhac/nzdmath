package ws

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHubRegisterAndBroadcast(t *testing.T) {
	h := NewHub()
	go h.Run()

	c1 := NewClient(1, nil, h)
	c2 := NewClient(2, nil, h)
	h.Register(c1)
	h.Register(c2)

	time.Sleep(10 * time.Millisecond)

	h.BroadcastMessage(map[string]interface{}{"id": 1})

	for _, c := range []*Client{c1, c2} {
		select {
		case payload := <-c.Send:
			var env Envelope
			if err := json.Unmarshal(payload, &env); err != nil {
				t.Fatalf("广播负载不是合法 JSON: %v", err)
			}
			if env.Type != TypeMessage {
				t.Fatalf("广播类型 = %q, want %q", env.Type, TypeMessage)
			}
		case <-time.After(time.Second):
			t.Fatal("1s 内未收到广播消息")
		}
	}
}

func TestHubUnregisterClosesSend(t *testing.T) {
	h := NewHub()
	go h.Run()

	c := NewClient(7, nil, h)
	h.Register(c)
	time.Sleep(10 * time.Millisecond)

	h.Unregister(c)
	time.Sleep(10 * time.Millisecond)

	// Unregister 后 Send 通道应被关闭
	if _, ok := <-c.Send; ok {
		t.Fatal("Unregister 后 Send 通道应已关闭")
	}
}

func TestHubBroadcastDoesNotBlockOnSlowClient(t *testing.T) {
	h := NewHub()
	go h.Run()

	slow := NewClient(1, nil, h)
	h.Register(slow)
	time.Sleep(10 * time.Millisecond)

	// 填满慢客户端的写缓冲
	for i := 0; i < SendBuffer; i++ {
		slow.Send <- []byte("x")
	}

	done := make(chan struct{})
	go func() {
		h.BroadcastPresence(3)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("慢客户端缓冲满时广播阻塞超时（应直接丢弃帧）")
	}
}

func TestHubBroadcastPresence(t *testing.T) {
	h := NewHub()
	go h.Run()

	c := NewClient(9, nil, h)
	h.Register(c)
	time.Sleep(10 * time.Millisecond)

	h.BroadcastPresence(5)

	select {
	case payload := <-c.Send:
		var env Envelope
		if err := json.Unmarshal(payload, &env); err != nil {
			t.Fatalf("presence 负载解析失败: %v", err)
		}
		if env.Type != TypePresence {
			t.Fatalf("presence 类型 = %q, want %q", env.Type, TypePresence)
		}
	case <-time.After(time.Second):
		t.Fatal("未收到 presence 广播")
	}
}
