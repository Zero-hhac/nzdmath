package model

import "time"

type DownloadLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	ResourceID uint      `gorm:"index;not null" json:"resource_id"`
	IP         string    `gorm:"type:varchar(64)" json:"ip"`
	UserAgent  string    `gorm:"type:varchar(500)" json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
	DeletedAt  *time.Time `gorm:"index" json:"-"`
}

func (DownloadLog) TableName() string {
	return "download_logs"
}