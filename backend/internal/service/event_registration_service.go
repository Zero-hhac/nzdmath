package service

import (
	"context"
	"errors"
	"fmt"
	"math-top/internal/dto"
	"math-top/internal/middleware"
	"math-top/internal/model"
	"math-top/internal/utils"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type EventRegistrationService struct {
	db       *gorm.DB
	rdb      *redis.Client
	notifier func(userIDs []uint) // WS 实时推送回调（报名成功 → 红点刷新）
}

func NewEventRegistrationService(db *gorm.DB, rdb *redis.Client) *EventRegistrationService {
	return &EventRegistrationService{db: db, rdb: rdb}
}

// SetNotifier 注入实时推送回调（router 组装时绑定 WebSocket Hub）
func (s *EventRegistrationService) SetNotifier(fn func(userIDs []uint)) {
	s.notifier = fn
}

// Register 报名活动：活动须已发布且未开始；容量满时拒绝。
// 报名成功后在同一事务内自动写入一条"报名成功"系统通知（顶栏红点提示）。
func (s *EventRegistrationService) Register(eventID, userID uint) error {
	var event model.Event
	if err := s.db.Where("id = ? AND status = ?", eventID, 1).First(&event).Error; err != nil {
		return errors.New("活动不存在或未发布")
	}
	if event.IsExpired {
		return errors.New("活动已过期，无法报名")
	}
	if time.Now().After(event.StartTime) {
		return errors.New("活动已开始，无法报名")
	}
	if event.Capacity > 0 {
		var count int64
		if err := s.db.Model(&model.EventRegistration{}).Where("event_id = ?", eventID).Count(&count).Error; err != nil {
			return errors.New("系统繁忙，请稍后再试")
		}
		if count >= int64(event.Capacity) {
			return errors.New("活动名额已满")
		}
	}
	location := event.Location
	if location == "" {
		location = "待定"
	}
	reg := model.EventRegistration{
		EventID:      eventID,
		UserID:       userID,
		Status:       model.RegStatusRegistered,
		RegisteredAt: time.Now(),
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&reg).Error; err != nil {
			return errors.New("报名失败，可能已报名过该活动")
		}
		notice := model.Notification{
			UserID: userID,
			Title:  "报名成功",
			Content: fmt.Sprintf("你已成功报名活动「%s」。\n活动时间：%s\n活动地点：%s\n请准时参加，现场凭姓名签到。",
				event.Title, event.StartTime.Format("2006-01-02 15:04"), location),
			Type: "activity",
		}
		if err := tx.Create(&notice).Error; err != nil {
			return errors.New("报名失败")
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 实时推送：报名成功通知的红点立即刷新
	if s.notifier != nil {
		s.notifier([]uint{userID})
	}
	return nil
}

// Cancel 取消报名（硬删记录，之后可重新报名）。
// 状态机约束：已签到、已开始或已过期的活动无法取消报名。
// 事务内同时写入"取消报名"通知，并实时推送红点。
func (s *EventRegistrationService) Cancel(eventID, userID uint) error {
	var reg model.EventRegistration
	if err := s.db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&reg).Error; err != nil {
		return errors.New("未报名该活动")
	}
	if reg.Status == model.RegStatusAttended {
		return errors.New("已签到的活动无法取消报名")
	}

	var event model.Event
	if err := s.db.First(&event, eventID).Error; err != nil {
		return errors.New("活动不存在")
	}
	if event.IsExpired {
		return errors.New("活动已结束，无法取消报名")
	}
	if time.Now().After(event.StartTime) {
		return errors.New("活动已开始，无法取消报名")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&reg).Error; err != nil {
			return errors.New("取消报名失败")
		}
		notice := model.Notification{
			UserID: userID,
			Title:  "取消报名通知",
			Content: fmt.Sprintf("你已成功取消活动「%s」的报名。\n如需重新参加，请在活动开始前重新报名。", event.Title),
			Type:   "activity",
		}
		if err := tx.Create(&notice).Error; err != nil {
			return errors.New("生成通知失败")
		}
		return nil
	})
	if err != nil {
		return err
	}

	if s.notifier != nil {
		s.notifier([]uint{userID})
	}
	return nil
}

func (s *EventRegistrationService) IsRegistered(eventID, userID uint) bool {
	var count int64
	s.db.Model(&model.EventRegistration{}).Where("event_id = ? AND user_id = ?", eventID, userID).Count(&count)
	return count > 0
}

func (s *EventRegistrationService) RegisteredCount(eventID uint) int64 {
	var count int64
	s.db.Model(&model.EventRegistration{}).Where("event_id = ?", eventID).Count(&count)
	return count
}

// MyRegistrations 会员的报名列表（含活动信息，按活动开始时间倒序）
func (s *EventRegistrationService) MyRegistrations(userID uint) ([]dto.EventRegistrationItem, error) {
	var regs []model.EventRegistration
	if err := s.db.Where("user_id = ?", userID).Order("id desc").Find(&regs).Error; err != nil {
		return nil, errors.New("获取报名列表失败")
	}
	if len(regs) == 0 {
		return []dto.EventRegistrationItem{}, nil
	}
	eventIDs := make([]uint, 0, len(regs))
	for _, r := range regs {
		eventIDs = append(eventIDs, r.EventID)
	}
	var events []model.Event
	s.db.Select("id, title, location, start_time, capacity").Where("id IN ?", eventIDs).Find(&events)
	em := make(map[uint]model.Event, len(events))
	for _, e := range events {
		em[e.ID] = e
	}
	out := make([]dto.EventRegistrationItem, 0, len(regs))
	for _, r := range regs {
		e, ok := em[r.EventID]
		if !ok {
			continue
		}
		out = append(out, dto.EventRegistrationItem{
			EventID:       r.EventID,
			EventTitle:    e.Title,
			EventLocation: e.Location,
			StartTime:     e.StartTime,
			Capacity:      e.Capacity,
			Status:        r.Status,
			RegisteredAt:  r.RegisteredAt,
			CheckedInAt:   r.CheckedInAt,
		})
	}
	return out, nil
}

// ListByEvent 管理员查看某活动报名名单（含用户信息）
func (s *EventRegistrationService) ListByEvent(eventID uint) ([]dto.EventRegistrationAdminItem, error) {
	var regs []model.EventRegistration
	if err := s.db.Where("event_id = ?", eventID).Order("registered_at asc, id asc").Find(&regs).Error; err != nil {
		return nil, errors.New("获取报名名单失败")
	}
	if len(regs) == 0 {
		return []dto.EventRegistrationAdminItem{}, nil
	}
	userIDs := make([]uint, 0, len(regs))
	for _, r := range regs {
		userIDs = append(userIDs, r.UserID)
	}
	var users []model.User
	s.db.Select("id, username, nickname, real_name, class_name, department").Where("id IN ?", userIDs).Find(&users)
	um := make(map[uint]model.User, len(users))
	for _, u := range users {
		um[u.ID] = u
	}
	out := make([]dto.EventRegistrationAdminItem, 0, len(regs))
	for _, r := range regs {
		u, ok := um[r.UserID]
		if !ok {
			continue
		}
		out = append(out, dto.EventRegistrationAdminItem{
			ID:           r.ID,
			UserID:       r.UserID,
			Username:     u.Username,
			Nickname:     u.Nickname,
			RealName:     u.RealName,
			ClassName:    u.ClassName,
			Department:   u.Department,
			Status:       r.Status,
			RegisteredAt: r.RegisteredAt,
			CheckedInAt:  r.CheckedInAt,
		})
	}
	return out, nil
}

// MarkAttended 签到
func (s *EventRegistrationService) MarkAttended(eventID, userID uint) error {
	var reg model.EventRegistration
	if err := s.db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&reg).Error; err != nil {
		return errors.New("该用户未报名该活动")
	}

	var event model.Event
	if err := s.db.First(&event, eventID).Error; err != nil {
		return errors.New("活动不存在")
	}

	now := time.Now()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.EventRegistration{}).
			Where("event_id = ? AND user_id = ?", eventID, userID).
			Updates(map[string]interface{}{"status": model.RegStatusAttended, "checked_in_at": now})
		if res.Error != nil {
			return res.Error
		}

		notice := model.Notification{
			UserID: userID,
			Title:  "签到成功",
			Content: fmt.Sprintf("你参加的活动「%s」已完成现场签到。\n签到时间：%s\n感谢你的积极参与！",
				event.Title,
				now.Format("2006-01-02 15:04:05"),
			),
			Type: "activity",
		}
		if err := tx.Create(&notice).Error; err != nil {
			return errors.New("生成通知失败")
		}
		return nil
	})
	if err != nil {
		return err
	}

	if s.notifier != nil {
		s.notifier([]uint{userID})
	}
	return nil
}

// CancelCheckin 取消签到（回到已报名状态）
func (s *EventRegistrationService) CancelCheckin(eventID, userID uint) error {
	var reg model.EventRegistration
	if err := s.db.Where("event_id = ? AND user_id = ?", eventID, userID).First(&reg).Error; err != nil {
		return errors.New("该用户未报名该活动")
	}

	var event model.Event
	if err := s.db.First(&event, eventID).Error; err != nil {
		return errors.New("活动不存在")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.EventRegistration{}).
			Where("event_id = ? AND user_id = ?", eventID, userID).
			Updates(map[string]interface{}{"status": model.RegStatusRegistered, "checked_in_at": nil})
		if res.Error != nil {
			return res.Error
		}

		notice := model.Notification{
			UserID: userID,
			Title:  "签到状态变更",
			Content: fmt.Sprintf("你参加的活动「%s」签到状态已被管理员撤销，当前恢复为「已报名」状态。",
				event.Title,
			),
			Type: "activity",
		}
		if err := tx.Create(&notice).Error; err != nil {
			return errors.New("生成通知失败")
		}
		return nil
	})
	if err != nil {
		return err
	}

	if s.notifier != nil {
		s.notifier([]uint{userID})
	}
	return nil
}

// AdminRemove 管理员移除报名记录
func (s *EventRegistrationService) AdminRemove(eventID, userID uint) error {
	res := s.db.Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&model.EventRegistration{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("报名记录不存在")
	}
	return nil
}

// Summary 后台报名管理页：全部活动 + 各自报名/签到人数
func (s *EventRegistrationService) Summary() ([]dto.EventRegistrationSummaryItem, error) {
	var events []model.Event
	if err := s.db.Select("id, title, start_time, capacity, status").Order("id desc").Find(&events).Error; err != nil {
		return nil, errors.New("获取活动列表失败")
	}
	var counts []struct {
		EventID uint
		Status  int
		Total   int64
	}
	if err := s.db.Model(&model.EventRegistration{}).
		Select("event_id, status, COUNT(*) as total").
		Group("event_id, status").
		Scan(&counts).Error; err != nil {
		return nil, errors.New("获取报名统计失败")
	}
	m := make(map[uint]map[int]int64, len(events))
	for _, c := range counts {
		if m[c.EventID] == nil {
			m[c.EventID] = make(map[int]int64)
		}
		m[c.EventID][c.Status] = c.Total
	}
	out := make([]dto.EventRegistrationSummaryItem, 0, len(events))
	for _, e := range events {
		registered := m[e.ID][model.RegStatusRegistered] + m[e.ID][model.RegStatusAttended]
		out = append(out, dto.EventRegistrationSummaryItem{
			ID:         e.ID,
			Title:      e.Title,
			StartTime:  e.StartTime,
			Capacity:   e.Capacity,
			Status:     e.Status,
			Registered: registered,
			Attended:   m[e.ID][model.RegStatusAttended],
		})
	}
	return out, nil
}

// BuildDetail 组装活动详情（含报名人数与当前用户报名状态）。
// tokenString 为空表示未登录。
func (s *EventRegistrationService) BuildDetail(event *model.Event, tokenString string) *dto.EventDetail {
	detail := &dto.EventDetail{
		ID:              event.ID,
		Title:           event.Title,
		Summary:         event.Summary,
		Content:         event.Content,
		Category:        event.Category,
		Location:        event.Location,
		StartTime:       event.StartTime,
		EndTime:         event.EndTime,
		CoverUrl:        event.CoverUrl,
		Capacity:        event.Capacity,
		IsExpired:       event.IsExpired,
		Status:          event.Status,
		IsFeatured:      event.IsFeatured,
		CreatedAt:       event.CreatedAt,
		RegisteredCount: s.RegisteredCount(event.ID),
	}
	if userID := s.resolveUserID(tokenString); userID > 0 {
		detail.IsRegistered = s.IsRegistered(event.ID, userID)
	}
	return detail
}

// resolveUserID 从 token 解析用户 ID（未登录返回 0），与 activity_service 同口径。
func (s *EventRegistrationService) resolveUserID(tokenString string) uint {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return 0
	}
	if s.rdb == nil {
		return 0
	}
	if s.rdb.Exists(context.Background(), middleware.UserTokenPrefix+tokenString).Val() == 0 {
		return 0
	}
	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		return 0
	}
	return claims.UserID
}
