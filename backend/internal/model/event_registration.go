package model

import "time"

// 报名状态
const (
	RegStatusRegistered = 1 // 已报名
	RegStatusAttended   = 2 // 已签到
)

type EventRegistration struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID      uint       `gorm:"not null;uniqueIndex:uk_event_user" json:"event_id"`
	UserID       uint       `gorm:"not null;uniqueIndex:uk_event_user" json:"user_id"`
	Status       int        `gorm:"type:tinyint;default:1" json:"status"` // 1-已报名 2-已签到
	RegisteredAt time.Time  `json:"registered_at"`
	CheckedInAt  *time.Time `json:"checked_in_at"`
	CreatedAt    time.Time  `json:"created_at"`
}
