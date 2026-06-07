package model

import "time"

type Comment struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	UserName   string    `gorm:"-" json:"user_name"`
	UserAvatar string    `gorm:"-" json:"user_avatar"`
	TargetType string    `gorm:"type:varchar(20);index;not null" json:"target_type"`
	TargetID   uint      `gorm:"index;not null" json:"target_id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Rating     int       `gorm:"type:tinyint;default:0" json:"rating"`
	ParentID   *uint     `gorm:"index" json:"parent_id"`
	LikeCount  int       `gorm:"type:int;default:0" json:"like_count"`
	ReplyCount int       `gorm:"type:int;default:0" json:"reply_count"`
	Status     int       `gorm:"type:tinyint;default:1;index" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  *time.Time `gorm:"index" json:"-"`
}

func (Comment) TableName() string {
	return "comments"
}

type CommentLike struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"index;not null;uniqueIndex:idx_user_comment" json:"user_id"`
	CommentID uint      `gorm:"index;not null;uniqueIndex:idx_user_comment" json:"comment_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (CommentLike) TableName() string {
	return "comment_likes"
}