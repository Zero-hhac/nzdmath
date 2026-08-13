package model

import "time"

type Notification struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint       `gorm:"not null;index:idx_user_read" json:"user_id"`
	Title     string     `gorm:"type:varchar(150);not null" json:"title"`
	Content   string     `gorm:"type:text" json:"content"`
	Type      string     `gorm:"type:varchar(30);default:'system'" json:"type"` // system/activity/reward 等
	IsRead    bool       `gorm:"type:boolean;default:false;index:idx_user_read" json:"is_read"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// NotificationBatch 一次发送记录（管理员后台的发送历史）
type NotificationBatch struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID   uint      `json:"admin_id"`
	Title     string    `gorm:"type:varchar(150);not null" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	Type      string    `gorm:"type:varchar(30);default:'system'" json:"type"`
	Target    string    `gorm:"type:varchar(255)" json:"target"` // 目标描述，如：全部会员 / 部门:宣传部 / 用户:3 人
	Count     int       `gorm:"default:0" json:"count"`
	CreatedAt time.Time `json:"created_at"`
}
