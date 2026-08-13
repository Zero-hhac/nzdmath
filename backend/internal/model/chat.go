package model

import "time"

type ChatMessage struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	UserName    string     `gorm:"-" json:"user_name"`
	UserAvatar  string     `gorm:"-" json:"user_avatar"`
	RealName    string     `gorm:"-" json:"real_name"`
	Department  string     `gorm:"-" json:"department"`
	MessageType string     `gorm:"type:varchar(20);not null;index" json:"message_type"`
	Content     string     `gorm:"type:text" json:"content"`
	FileName    string     `gorm:"type:varchar(255)" json:"file_name"`
	FileURL     string     `gorm:"type:varchar(500)" json:"file_url"`
	FilePath    string     `gorm:"type:varchar(500)" json:"-"`
	FileSize    int64      `json:"file_size"`
	FileExt     string     `gorm:"type:varchar(20)" json:"file_ext"`
	FileMime    string     `gorm:"type:varchar(100)" json:"file_mime"`
	CreatedAt   time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"-"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

type ChatPresence struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	IsOnline   bool      `gorm:"type:boolean;default:false;index" json:"is_online"`
	JoinedAt   time.Time `json:"joined_at"`
	LastSeenAt time.Time `gorm:"index" json:"last_seen_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (ChatPresence) TableName() string {
	return "chat_presences"
}
