package model

import "time"

type Admin struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"type:varchar(50);not null;uniqueIndex" json:"username"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string     `gorm:"type:varchar(50)" json:"nickname"`
	Email        string     `gorm:"type:varchar(100)" json:"email"`
	Role         string     `gorm:"type:varchar(20);default:'admin'" json:"role"`
	Status       int        `gorm:"type:tinyint;default:1" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (Admin) TableName() string {
	return "admins"
}
