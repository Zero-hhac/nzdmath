package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string    `gorm:"type:varchar(50)" json:"nickname"`
	Email        *string    `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	Avatar       string    `gorm:"type:varchar(500)" json:"avatar"`
	Bio          string    `gorm:"type:text" json:"bio"`
	Role         string    `gorm:"type:varchar(20);default:'member';index" json:"role"`
	Status       int       `gorm:"type:tinyint;default:1;index" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}