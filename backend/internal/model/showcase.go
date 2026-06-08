package model

import "time"

type Showcase struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"type:varchar(200);not null" json:"title"`           // 作品标题
	Author      string    `gorm:"type:varchar(100);not null" json:"author"`          // 作者
	Field       string    `gorm:"type:varchar(50)" json:"field"`                     // 领域：数学建模/统计分析/算法竞赛等
	Competition string    `gorm:"type:varchar(100)" json:"competition"`              // 比赛名称
	Summary     string    `gorm:"type:text" json:"summary"`                          // 作品简介
	CoverURL    string    `gorm:"type:varchar(500)" json:"cover_url"`                // 封面图链接
	H5URL       string    `gorm:"type:varchar(500)" json:"h5_url"`                   // H5网页链接
	ViewCount   int       `gorm:"type:int;default:0" json:"view_count"`              // 浏览次数
	Status      int       `gorm:"type:tinyint;default:1" json:"status"`              // 状态：1-发布，0-草稿
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Showcase) TableName() string {
	return "showcases"
}


