package model

import "time"

type Event struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title      string    `gorm:"type:varchar(150);not null" json:"title"`       // 活动标题
	Summary    string    `gorm:"type:varchar(255)" json:"summary"`              // 摘要
	Content    string    `gorm:"type:longtext" json:"content"`                  // 详情内容
	Category   string    `gorm:"type:varchar(50)" json:"category"`              // 分类
	Location   string    `gorm:"type:varchar(255)" json:"location"`             // 活动地点
	StartTime  time.Time `json:"start_time"`                                    // 开始时间
	EndTime    time.Time `json:"end_time"`                                      // 结束时间
	CoverUrl   string    `gorm:"type:varchar(255)" json:"cover_url"`            // 封面图
	Capacity   int       `gorm:"type:int;default:0" json:"capacity"`            // 报名名额上限，0 表示不限
	IsExpired  bool      `gorm:"type:boolean;default:false" json:"is_expired"`  // 管理员手动标记：活动已过期（公开列表下线、禁止报名）
	Status     int       `gorm:"type:tinyint;default:1" json:"status"`          // 状态：1-发布，0-草稿
	IsFeatured bool      `gorm:"type:boolean;default:false" json:"is_featured"` // 是否推荐到首页
	CreatedBy  uint      `json:"created_by"`                                    // 后台管理员ID
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
