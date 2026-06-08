package dto

import "time"

// AdminLoginRequest 管理员登录
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminListQuery 通用列表查询参数
type AdminListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   *int   `form:"status"`
	Keyword  string `form:"keyword"`
}

// AdminCreateEventRequest 创建活动
type AdminCreateEventRequest struct {
	Title      string    `json:"title" binding:"required,min=2,max=150"`
	Summary    string    `json:"summary" binding:"max=255"`
	Content    string    `json:"content"`
	Category   string    `json:"category"`
	Location   string    `json:"location"`
	StartTime  time.Time `json:"start_time" binding:"required"`
	EndTime    time.Time `json:"end_time" binding:"required"`
	CoverURL   string    `json:"cover_url"`
	Status     int       `json:"status" binding:"oneof=0 1"`
	IsFeatured bool      `json:"is_featured"`
}

// AdminUpdateEventRequest 更新活动
type AdminUpdateEventRequest = AdminCreateEventRequest

// AdminCreateNewsRequest 创建资讯
type AdminCreateNewsRequest struct {
	Title      string `json:"title" binding:"required,min=2,max=200"`
	Summary    string `json:"summary"`
	Content    string `json:"content"`
	Category   string `json:"category"`
	Tag        string `json:"tag"`
	CoverURL   string `json:"cover_url"`
	Status     int    `json:"status" binding:"oneof=0 1"`
	IsFeatured bool   `json:"is_featured"`
}

// AdminUpdateNewsRequest 更新资讯
type AdminUpdateNewsRequest = AdminCreateNewsRequest

// AdminUpdateResourceRequest 更新资源（资源无需 POST，文件上传走公开接口）
type AdminUpdateResourceRequest struct {
	Title      string `json:"title" binding:"required,min=2"`
	Summary    string `json:"summary"`
	Category   string `json:"category"`
	CoverURL   string `json:"cover_url"`
	Status     int    `json:"status" binding:"oneof=0 1"`
	IsFeatured bool   `json:"is_featured"`
}

// AdminCreateShowcaseRequest 创建作品
type AdminCreateShowcaseRequest struct {
	Title       string `json:"title" binding:"required,min=2,max=200"`
	Author      string `json:"author" binding:"required"`
	Field       string `json:"field"`
	Competition string `json:"competition"`
	Summary     string `json:"summary"`
	CoverURL    string `json:"cover_url"`
	H5URL       string `json:"h5_url"`
	Status      int    `json:"status" binding:"oneof=0 1"`
}


// AdminUpdateShowcaseRequest 更新作品
type AdminUpdateShowcaseRequest = AdminCreateShowcaseRequest

// AdminDashboardResponse 后台首页聚合数据
type AdminDashboardResponse struct {
	Counts        DashboardCounts   `json:"counts"`
	TodayNew      DashboardTodayNew `json:"today_new"`
	Trend         DashboardTrend    `json:"trend_7days"`
	TodayActivity TodayActivity     `json:"today_activity"` // 新增今日实时流量数据
	Activity      ActivityTrend     `json:"activity_trend"`   // 新增7天流量趋势
}

type TodayActivity struct {
	PV  int64 `json:"pv"`
	UV  int64 `json:"uv"`
	DAU int64 `json:"dau"`
}

type ActivityTrend struct {
	Dates []string `json:"dates"`
	PV    []int64  `json:"pv"`
	UV    []int64  `json:"uv"`
	DAU   []int64  `json:"dau"`
}

type DashboardCounts struct {
	Users     int64 `json:"users"`
	Events    int64 `json:"events"`
	News      int64 `json:"news"`
	Resources int64 `json:"resources"`
	Showcases int64 `json:"showcases"`
}

type DashboardTodayNew struct {
	Events int64 `json:"events"`
	News   int64 `json:"news"`
}

type DashboardTrend struct {
	Dates  []string `json:"dates"`
	Events []int64  `json:"events"`
	News   []int64  `json:"news"`
}
