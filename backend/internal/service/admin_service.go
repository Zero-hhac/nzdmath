package service

import (
	"context"
	"errors"
	"fmt"
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

// EnsureDefaultAdmin 若 admins 表为空，创建默认账号 admin。
// M4：任何模式（含本地 debug）都必须通过 ADMIN_INITIAL_PASSWORD 设置初始密码，
// 彻底杜绝 admin/admin123 默认口令。
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
		return errors.New("必须通过 ADMIN_INITIAL_PASSWORD 环境变量设置初始管理员密码")
	}
	// 与会员同一套强度策略：6-72 位且必须含字母和数字
	if err := utils.ValidatePasswordStrength(password); err != nil {
		return err
	}
	switch password {
	case "admin123", "password123", "123456", "admin", "password":
		return errors.New("初始管理员密码过弱，请使用高强度密码")
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

	// #5：admin_token 有效期读取配置，并登记用户→token 索引便于统一吊销
	ttl := time.Duration(config.GlobalConfig.JWT.ExpireHours) * time.Hour
	redisKey := middleware.AdminTokenPrefix + token
	if err := s.rdb.Set(context.Background(), redisKey, admin.ID, ttl).Err(); err != nil {
		return "", nil, errors.New("Redis 保存 token 失败")
	}
	IndexUserToken(s.rdb, admin.ID, token, middleware.AdminTokenPrefix, ttl)

	now := time.Now()
	s.db.Model(&admin).Update("last_login_at", &now)

	return token, &admin, nil
}

func (s *AdminService) Logout(token string) error {
	redisKey := middleware.AdminTokenPrefix + token
	if err := s.rdb.Del(context.Background(), redisKey).Err(); err != nil {
		return err
	}
	// #5：登出时同步从用户索引移除
	if claims, err := utils.ParseToken(token); err == nil {
		UnindexUserToken(s.rdb, claims.UserID, token, middleware.AdminTokenPrefix)
	}
	return nil
}

// ChangePassword 管理员自助修改密码（后台“账号设置”入口）。
// 新密码与会员同一强度策略；校验原密码后更新哈希。
func (s *AdminService) ChangePassword(adminID uint, oldPassword, newPassword string) error {
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return errors.New("管理员不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	if err := s.db.Model(&admin).Update("password_hash", string(hashed)).Error; err != nil {
		return errors.New("修改密码失败")
	}
	// 同步更新会员用户表
	s.db.Model(&model.User{}).Where("username = ?", admin.Username).Update("password_hash", string(hashed))
	// #5：改密后吊销该管理员全部会话（admin_token 与同名会员行 user_token 一并吊销）
	RevokeUserTokens(s.rdb, admin.ID)
	var shadow model.User
	if err := s.db.Where("username = ?", admin.Username).First(&shadow).Error; err == nil {
		RevokeUserTokens(s.rdb, shadow.ID)
	}
	return nil
}

// SyncDailyMetrics 针对过去 6 天的历史指标，若数据库无归档则从 Redis 提取并保存，
// 写入后清除 Redis Key 以释放内存。Dashboard 调用时做兜底，router 中的常驻定时任务
// 定期执行，防止管理员长时间不登录导致 Redis key 过期、数据永久丢失。
func (s *AdminService) SyncDailyMetrics(ctx context.Context) error {
	now := time.Now()
	for i := 6; i >= 1; i-- {
		dStr := now.AddDate(0, 0, -i).Format("2006-01-02")

		var count int64
		if err := s.db.Model(&model.DailyMetric{}).Where("date = ?", dStr).Count(&count).Error; err != nil {
			return fmt.Errorf("检查 %s 归档状态失败: %w", dStr, err)
		}
		if count > 0 {
			continue
		}

		// 从 Redis 中读取
		pvKey := "dau:pv:" + dStr
		uvKey := "dau:ip:" + dStr
		dauKey := "dau:user:" + dStr

		pvValStr, err := s.rdb.Get(ctx, pvKey).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return fmt.Errorf("读取 %s PV 失败: %w", dStr, err)
		}
		var pvVal int64
		if pvValStr != "" {
			if pvVal, err = strconv.ParseInt(pvValStr, 10, 64); err != nil {
				return fmt.Errorf("解析 %s PV 失败: %w", dStr, err)
			}
		}

		uvVal, err := s.rdb.PFCount(ctx, uvKey).Result()
		if err != nil {
			return fmt.Errorf("读取 %s UV 失败: %w", dStr, err)
		}
		dauVal, err := s.rdb.PFCount(ctx, dauKey).Result()
		if err != nil {
			return fmt.Errorf("读取 %s DAU 失败: %w", dStr, err)
		}

		// 写入 MySQL 归档
		metric := model.DailyMetric{
			Date:      dStr,
			PV:        pvVal,
			UV:        uvVal,
			DAU:       dauVal,
			CreatedAt: now,
		}
		if err := s.db.Create(&metric).Error; err != nil {
			return fmt.Errorf("归档 %s 指标失败: %w", dStr, err)
		}
		// 成功写入后删除 Redis key 以释放内存
		if err := s.rdb.Del(ctx, pvKey, uvKey, dauKey).Err(); err != nil {
			slog.Warn("清理指标 Redis key 失败", "date", dStr, "err", err)
		}
	}
	return nil
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

	if err := s.db.Model(&model.User{}).Count(&resp.Counts.Users).Error; err != nil {
		return nil, fmt.Errorf("统计用户数失败: %w", err)
	}
	if err := s.db.Model(&model.Event{}).Count(&resp.Counts.Events).Error; err != nil {
		return nil, fmt.Errorf("统计活动数失败: %w", err)
	}
	if err := s.db.Model(&model.News{}).Count(&resp.Counts.News).Error; err != nil {
		return nil, fmt.Errorf("统计新闻数失败: %w", err)
	}
	if err := s.db.Model(&model.Resource{}).Count(&resp.Counts.Resources).Error; err != nil {
		return nil, fmt.Errorf("统计资源数失败: %w", err)
	}
	if err := s.db.Model(&model.Showcase{}).Count(&resp.Counts.Showcases).Error; err != nil {
		return nil, fmt.Errorf("统计作品数失败: %w", err)
	}

	// 本地时区零点：与 activity_service 按本地日期生成 Redis key 的口径保持一致。
	// 注意不能用 time.Now().Truncate(24*time.Hour)，那会按 UTC 零点截断，东八区
	// 每天 00:00-08:00 的“今日”数据会被算错天。
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if err := s.db.Model(&model.Event{}).Where("created_at >= ?", today).Count(&resp.TodayNew.Events).Error; err != nil {
		return nil, fmt.Errorf("统计今日新增活动失败: %w", err)
	}
	if err := s.db.Model(&model.News{}).Where("created_at >= ?", today).Count(&resp.TodayNew.News).Error; err != nil {
		return nil, fmt.Errorf("统计今日新增新闻失败: %w", err)
	}

	type DateCount struct {
		Day   string `gorm:"column:day"`
		Total int64  `gorm:"column:total"`
	}
	sevenDaysAgo := today.AddDate(0, 0, -6)

	var eventCounts []DateCount
	if err := s.db.Model(&model.Event{}).
		Select("DATE(created_at) as day, COUNT(*) as total").
		Where("created_at >= ?", sevenDaysAgo).
		Group("DATE(created_at)").
		Scan(&eventCounts).Error; err != nil {
		return nil, fmt.Errorf("统计活动趋势失败: %w", err)
	}
	eventMap := make(map[string]int64, len(eventCounts))
	for _, c := range eventCounts {
		eventMap[c.Day] = c.Total
	}

	var newsCounts []DateCount
	if err := s.db.Model(&model.News{}).
		Select("DATE(created_at) as day, COUNT(*) as total").
		Where("created_at >= ?", sevenDaysAgo).
		Group("DATE(created_at)").
		Scan(&newsCounts).Error; err != nil {
		return nil, fmt.Errorf("统计新闻趋势失败: %w", err)
	}
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
	// 4. 同步前几天的指标数据（失败不阻断 Dashboard，由常驻定时任务兜底重试）
	if err := s.SyncDailyMetrics(ctx); err != nil {
		slog.Warn("同步每日指标失败", "err", err)
	}

	// 5. 统计今日流量实时数据
	todayStr := time.Now().Format("2006-01-02")
	pvTodayStr, err := s.rdb.Get(ctx, "dau:pv:"+todayStr).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("读取今日 PV 失败: %w", err)
	}
	var pvToday int64
	if pvTodayStr != "" {
		if pvToday, err = strconv.ParseInt(pvTodayStr, 10, 64); err != nil {
			return nil, fmt.Errorf("解析今日 PV 失败: %w", err)
		}
	}
	uvToday, err := s.rdb.PFCount(ctx, "dau:ip:"+todayStr).Result()
	if err != nil {
		return nil, fmt.Errorf("读取今日 UV 失败: %w", err)
	}
	dauToday, err := s.rdb.PFCount(ctx, "dau:user:"+todayStr).Result()
	if err != nil {
		return nil, fmt.Errorf("读取今日 DAU 失败: %w", err)
	}

	resp.TodayActivity = dto.TodayActivity{
		PV:  pvToday,
		UV:  uvToday,
		DAU: dauToday,
	}

	// Global PV uses a counter and global UV uses one HyperLogLog. Unlike
	// daily UV, the global set does not double-count visitors across days.
	totalPV, err := s.rdb.Get(ctx, "dau:pv:all").Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("读取累计 PV 失败: %w", err)
	}
	if errors.Is(err, redis.Nil) {
		var archivedPV int64
		if err := s.db.Model(&model.DailyMetric{}).Select("COALESCE(SUM(pv), 0)").Scan(&archivedPV).Error; err != nil {
			return nil, fmt.Errorf("统计归档 PV 失败: %w", err)
		}
		totalPV = archivedPV + pvToday
		if err := s.rdb.SetNX(ctx, "dau:pv:all", totalPV, 0).Err(); err != nil {
			return nil, fmt.Errorf("初始化累计 PV 失败: %w", err)
		}
	}
	totalUV, err := s.rdb.PFCount(ctx, "dau:ip:all").Result()
	if err != nil {
		return nil, fmt.Errorf("读取累计 UV 失败: %w", err)
	}
	if totalUV == 0 {
		// Keep the current day's visitors visible immediately after deployment.
		todayUV, err := s.rdb.PFCount(ctx, "dau:ip:"+todayStr).Result()
		if err != nil {
			return nil, fmt.Errorf("读取今日 UV 失败: %w", err)
		}
		if todayUV > 0 {
			if err := s.rdb.PFMerge(ctx, "dau:ip:all", "dau:ip:"+todayStr).Err(); err != nil {
				return nil, fmt.Errorf("合并累计 UV 失败: %w", err)
			}
			if totalUV, err = s.rdb.PFCount(ctx, "dau:ip:all").Result(); err != nil {
				return nil, fmt.Errorf("读取累计 UV 失败: %w", err)
			}
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
	if err := s.db.Where("date >= ?", time.Now().AddDate(0, 0, -6).Format("2006-01-02")).
		Where("date < ?", todayStr).
		Order("date asc").
		Find(&historyMetrics).Error; err != nil {
		return nil, fmt.Errorf("读取历史指标失败: %w", err)
	}

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
