package model

import "time"

type DailyMetric struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Date      string    `gorm:"type:varchar(10);uniqueIndex;not null" json:"date"` // 格式 YYYY-MM-DD
	DAU       int64     `gorm:"type:bigint;default:0" json:"dau"`
	UV        int64     `gorm:"type:bigint;default:0" json:"uv"`
	PV        int64     `gorm:"type:bigint;default:0" json:"pv"`
	CreatedAt time.Time `json:"created_at"`
}

func (DailyMetric) TableName() string {
	return "daily_metrics"
}
