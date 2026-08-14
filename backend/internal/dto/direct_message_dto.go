package dto

import "time"

type DirectMessageItem struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	SenderType   string    `json:"sender_type"`
	AdminID      uint      `json:"admin_id"`
	MessageType  string    `json:"message_type"`
	Content      string    `json:"content"`
	FileName     string    `json:"file_name"`
	FileURL      string    `json:"file_url"`
	FileSize     int64     `json:"file_size"`
	FileExt      string    `json:"file_ext"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
	SenderName   string    `json:"sender_name"`
	SenderAvatar string    `json:"sender_avatar"`
}

type DirectConversationItem struct {
	UserID          uint      `json:"user_id"`
	Username        string    `json:"username"`
	Nickname        string    `json:"nickname"`
	RealName        string    `json:"real_name"`
	ClassName       string    `json:"class_name"`
	Department      string    `json:"department"`
	Avatar          string    `json:"avatar"`
	LastMessage     string    `json:"last_message"`
	LastMessageType string    `json:"last_message_type"`
	LastMessageAt   time.Time `json:"last_message_at"`
	UnreadCount     int64     `json:"unread_count"`
}
