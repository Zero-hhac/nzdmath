package service

import (
	"context"
	"errors"
	"log/slog"
	"math-top/internal/config"
	"math-top/internal/dto"
	"math-top/internal/middleware"
	"math-top/internal/model"
	"math-top/internal/utils"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewAdminService(db *gorm.DB, rdb *redis.Client) *AdminService {
	return &AdminService{db: db, rdb: rdb}
}

// EnsureDefaultAdmin 若 admins 表为空，创建默认账号 admin/admin123
func (s *AdminService) EnsureDefaultAdmin() error {
	var count int64
	if err := s.db.Model(&model.Admin{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	password := os.Getenv("ADMIN_INITIAL_PASSWORD")
	if password == "" {
		if config.GlobalConfig.App.Mode == "release" {
			return errors.New("生产环境必须通过 ADMIN_INITIAL_PASSWORD 设置初始管理员密码")
		}
		password = "admin123"
		slog.Warn("开发模式创建默认管理员 admin/admin123，生产环境请设置 ADMIN_INITIAL_PASSWORD")
	}
	if len(password) < 6 {
		return errors.New("初始管理员密码长度不能少于 6 位")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now()
	admin := model.Admin{
		Username:     "admin",
		PasswordHash: string(hashed),
		Nickname:     "超级管理员",
		Email:        "admin@math-top.com",
		Role:         "super_admin",
		Status:       1,
		LastLoginAt:  &now,
	}
	return s.db.Create(&admin).Error
}

func (s *AdminService) Login(username, password string) (string, *model.Admin, error) {
	if username == "" || password == "" {
		return "", nil, errors.New("用户名或密码不能为空")
	}
	var admin model.Admin
	if err := s.db.Where("username = ?", username).First(&admin).Error; err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	if admin.Status == 0 {
		return "", nil, errors.New("账号已被禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := utils.GenerateToken(admin.ID, admin.Username, admin.Role)
	if err != nil {
		return "", nil, errors.New("生成 Token 失败")
	}

	redisKey := middleware.AdminTokenPrefix + token
	if err := s.rdb.Set(context.Background(), redisKey, admin.ID, 24*time.Hour).Err(); err != nil {
		return "", nil, errors.New("Redis 保存 token 失败")
	}

	now := time.Now()
	s.db.Model(&admin).Update("last_login_at", &now)

	return token, &admin, nil
}

func (s *AdminService) Logout(token string) error {
	redisKey := middleware.AdminTokenPrefix + token
	return s.rdb.Del(context.Background(), redisKey).Err()
}

// GetDashboard 聚合统计数据
// syncDailyMetrics 针对过去 6 天的历史指标，若数据库无归档则从 Redis 提取并保存，写入后清除 Redis Key 以释放内存
func (s *AdminService) syncDailyMetrics(ctx context.Context) {
	now := time.Now()
	for i := 6; i >= 1; i-- {
		dStr := now.AddDate(0, 0, -i).Format("2006-01-02")

		var count int64
		s.db.Model(&model.DailyMetric{}).Where("date = ?", dStr).Count(&count)
		if count > 0 {
			continue
		}

		// 从 Redis 中读取
		pvKey := "dau:pv:" + dStr
		uvKey := "dau:ip:" + dStr
		dauKey := "dau:user:" + dStr

		pvValStr, _ := s.rdb.Get(ctx, pvKey).Result()
		var pvVal int64
		if pvValStr != "" {
			pvVal, _ = strconv.ParseInt(pvValStr, 10, 64)
		}

		uvVal := s.rdb.PFCount(ctx, uvKey).Val()
		dauVal := s.rdb.PFCount(ctx, dauKey).Val()

		// 写入 MySQL 归档
		metric := model.DailyMetric{
			Date:      dStr,
			PV:        pvVal,
			UV:        uvVal,
			DAU:       dauVal,
			CreatedAt: now,
		}
		if err := s.db.Create(&metric).Error; err == nil {
			// 成功写入后删除 Redis key 以释放内存
			s.rdb.Del(ctx, pvKey, uvKey, dauKey)
		}
	}
}

// GetDashboard 聚合统计数据
func (s *AdminService) GetDashboard() (*dto.AdminDashboardResponse, error) {
	resp := &dto.AdminDashboardResponse{
		Trend: dto.DashboardTrend{
			Dates:  []string{},
			Events: []int64{},
			News:   []int64{},
		},
	}

	s.db.Model(&model.User{}).Count(&resp.Counts.Users)
	s.db.Model(&model.Event{}).Count(&resp.Counts.Events)
	s.db.Model(&model.News{}).Count(&resp.Counts.News)
	s.db.Model(&model.Resource{}).Count(&resp.Counts.Resources)
	s.db.Model(&model.Showcase{}).Count(&resp.Counts.Showcases)

	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&model.Event{}).Where("created_at >= ?", today).Count(&resp.TodayNew.Events)
	s.db.Model(&model.News{}).Where("created_at >= ?", today).Count(&resp.TodayNew.News)

	type DateCount struct {
		Day   string `gorm:"column:day"`
		Total int64  `gorm:"column:total"`
	}
	sevenDaysAgo := today.AddDate(0, 0, -6)

	var eventCounts []DateCount
	s.db.Model(&model.Event{}).
		Select("DATE(created_at) as day, COUNT(*) as total").
		Where("created_at >= ?", sevenDaysAgo).
		Group("DATE(created_at)").
		Scan(&eventCounts)
	eventMap := make(map[string]int64, len(eventCounts))
	for _, c := range eventCounts {
		eventMap[c.Day] = c.Total
	}

	var newsCounts []DateCount
	s.db.Model(&model.News{}).
		Select("DATE(created_at) as day, COUNT(*) as total").
		Where("created_at >= ?", sevenDaysAgo).
		Group("DATE(created_at)").
		Scan(&newsCounts)
	newsMap := make(map[string]int64, len(newsCounts))
	for _, c := range newsCounts {
		newsMap[c.Day] = c.Total
	}

	for i := 6; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")
		resp.Trend.Dates = append(resp.Trend.Dates, dateStr)
		resp.Trend.Events = append(resp.Trend.Events, eventMap[dateStr])
		resp.Trend.News = append(resp.Trend.News, newsMap[dateStr])
	}

	ctx := context.Background()
	// 4. 同步前几天的指标数据
	s.syncDailyMetrics(ctx)

	// 5. 统计今日流量实时数据
	todayStr := time.Now().Format("2006-01-02")
	pvTodayStr, _ := s.rdb.Get(ctx, "dau:pv:"+todayStr).Result()
	var pvToday int64
	if pvTodayStr != "" {
		pvToday, _ = strconv.ParseInt(pvTodayStr, 10, 64)
	}
	uvToday := s.rdb.PFCount(ctx, "dau:ip:"+todayStr).Val()
	dauToday := s.rdb.PFCount(ctx, "dau:user:"+todayStr).Val()

	resp.TodayActivity = dto.TodayActivity{
		PV:  pvToday,
		UV:  uvToday,
		DAU: dauToday,
	}

	// Global PV uses a counter and global UV uses one HyperLogLog. Unlike
	// daily UV, the global set does not double-count visitors across days.
	totalPV, err := s.rdb.Get(ctx, "dau:pv:all").Int64()
	if err == redis.Nil {
		var archivedPV int64
		s.db.Model(&model.DailyMetric{}).Select("COALESCE(SUM(pv), 0)").Scan(&archivedPV)
		totalPV = archivedPV + pvToday
		_ = s.rdb.SetNX(ctx, "dau:pv:all", totalPV, 0).Err()
	}
	totalUV := s.rdb.PFCount(ctx, "dau:ip:all").Val()
	if totalUV == 0 {
		// Keep the current day's visitors visible immediately after deployment.
		if todayUV := s.rdb.PFCount(ctx, "dau:ip:"+todayStr).Val(); todayUV > 0 {
			s.rdb.PFMerge(ctx, "dau:ip:all", "dau:ip:"+todayStr)
			totalUV = s.rdb.PFCount(ctx, "dau:ip:all").Val()
		}
	}
	resp.TotalActivity = dto.TotalActivity{PV: totalPV, UV: totalUV}

	// 6. 构建 7 天活跃与流量趋势数据（6天历史归档 + 1天今日实时）
	resp.Activity = dto.ActivityTrend{
		Dates: []string{},
		PV:    []int64{},
		UV:    []int64{},
		DAU:   []int64{},
	}

	var historyMetrics []model.DailyMetric
	s.db.Where("date >= ?", time.Now().AddDate(0, 0, -6).Format("2006-01-02")).
		Where("date < ?", todayStr).
		Order("date asc").
		Find(&historyMetrics)

	historyMap := make(map[string]model.DailyMetric)
	for _, m := range historyMetrics {
		historyMap[m.Date] = m
	}

	for i := 6; i >= 0; i-- {
		dayStr := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		resp.Activity.Dates = append(resp.Activity.Dates, dayStr)

		if i == 0 {
			resp.Activity.PV = append(resp.Activity.PV, pvToday)
			resp.Activity.UV = append(resp.Activity.UV, uvToday)
			resp.Activity.DAU = append(resp.Activity.DAU, dauToday)
		} else {
			if val, ok := historyMap[dayStr]; ok {
				resp.Activity.PV = append(resp.Activity.PV, val.PV)
				resp.Activity.UV = append(resp.Activity.UV, val.UV)
				resp.Activity.DAU = append(resp.Activity.DAU, val.DAU)
			} else {
				resp.Activity.PV = append(resp.Activity.PV, 0)
				resp.Activity.UV = append(resp.Activity.UV, 0)
				resp.Activity.DAU = append(resp.Activity.DAU, 0)
			}
		}
	}

	return resp, nil
}
