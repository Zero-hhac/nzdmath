package model

import "time"

type News struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string     `gorm:"type:varchar(200);not null" json:"title"`       // 资讯标题
	Summary     string     `gorm:"type:text" json:"summary"`                      // 摘要
	Content     string     `gorm:"type:longtext" json:"content"`                  // 详情内容
	Category    string     `gorm:"type:varchar(50)" json:"category"`              // 分类
	Tag         string     `gorm:"type:varchar(100)" json:"tag"`                  // 标签
	CoverURL    string     `gorm:"type:varchar(500)" json:"cover_url"`            // 封面图链接
	Status      int        `gorm:"type:tinyint;default:1" json:"status"`          // 状态：1-发布，0-草稿
	IsFeatured  bool       `gorm:"type:boolean;default:false" json:"is_featured"` // 是否推荐到首页
	PublishedAt *time.Time `json:"published_at"`                                  // 发布时间
	CreatedBy   uint       `json:"created_by"`                                    // 后台管理员ID
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (News) TableName() string {
	return "news"
}
