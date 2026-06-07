package model

import "time"

type Favorite struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`                // 用户ID
	TargetID   uint      `gorm:"not null;index" json:"target_id"`              // 文章/资源ID
	TargetType string    `gorm:"type:varchar(50);not null" json:"target_type"` // 目标类型：resource/news/showcase/event
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}
