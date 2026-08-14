package model

import "time"

type DirectMessage struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	SenderType  string     `gorm:"type:varchar(10);not null;index" json:"sender_type"` // "user" 或 "admin"
	AdminID     uint       `gorm:"default:0" json:"admin_id"`
	MessageType string     `gorm:"type:varchar(20);not null;default:'text'" json:"message_type"` // "text", "image", "file"
	Content     string     `gorm:"type:text" json:"content"`
	FileName    string     `gorm:"type:varchar(255)" json:"file_name"`
	FileURL     string     `gorm:"type:varchar(500)" json:"file_url"`
	FilePath    string     `gorm:"type:varchar(500)" json:"-"`
	FileSize    int64      `json:"file_size"`
	FileExt     string     `gorm:"type:varchar(20)" json:"file_ext"`
	FileMime    string     `gorm:"type:varchar(100)" json:"file_mime"`
	IsRead      bool       `gorm:"default:false;index" json:"is_read"`
	CreatedAt   time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"-"`
}

func (DirectMessage) TableName() string {
	return "direct_messages"
}
