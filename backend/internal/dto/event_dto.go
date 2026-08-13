package dto

import "time"

// EventDetail 活动详情（内嵌 model.Event，字段平铺；附加报名信息）
type EventDetail struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary"`
	Content         string    `json:"content"`
	Category        string    `json:"category"`
	Location        string    `json:"location"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	CoverUrl        string    `json:"cover_url"`
	Capacity        int       `json:"capacity"`
	IsExpired       bool      `json:"is_expired"`
	Status          int       `json:"status"`
	IsFeatured      bool      `json:"is_featured"`
	CreatedAt       time.Time `json:"created_at"`
	RegisteredCount int64     `json:"registered_count"` // 已报名人数
	IsRegistered    bool      `json:"is_registered"`    // 当前用户是否已报名
}

// EventRegistrationItem 会员侧"我的报名"列表项
type EventRegistrationItem struct {
	EventID       uint       `json:"event_id"`
	EventTitle    string     `json:"event_title"`
	EventLocation string     `json:"event_location"`
	StartTime     time.Time  `json:"start_time"`
	Capacity      int        `json:"capacity"`
	Status        int        `json:"status"` // 1-已报名 2-已签到
	RegisteredAt  time.Time  `json:"registered_at"`
	CheckedInAt   *time.Time `json:"checked_in_at"`
}

// EventRegistrationAdminItem 管理员侧报名名单项
type EventRegistrationAdminItem struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"user_id"`
	Username     string     `json:"username"`
	Nickname     string     `json:"nickname"`
	RealName     string     `json:"real_name"`
	ClassName    string     `json:"class_name"`
	Department   string     `json:"department"`
	Status       int        `json:"status"`
	RegisteredAt time.Time  `json:"registered_at"`
	CheckedInAt  *time.Time `json:"checked_in_at"`
}

// EventRegistrationSummaryItem 后台报名管理页的活动汇总项
type EventRegistrationSummaryItem struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	StartTime  time.Time `json:"start_time"`
	Capacity   int       `json:"capacity"`
	Status     int       `json:"status"`
	Registered int64     `json:"registered"` // 已报名（含已签到）
	Attended   int64     `json:"attended"`   // 已签到
}
