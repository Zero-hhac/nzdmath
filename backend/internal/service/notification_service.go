package service

import (
	"errors"
	"fmt"
	"math-top/internal/consts"
	"math-top/internal/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

type NotificationService struct {
	db       *gorm.DB
	notifier func(userIDs []uint) // WS 实时推送回调（新通知 → 在线用户红点刷新）
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// SetNotifier 注入实时推送回调（router 组装时绑定 WebSocket Hub）
func (s *NotificationService) SetNotifier(fn func(userIDs []uint)) {
	s.notifier = fn
}

// NotificationTarget 发送目标：all=全部会员 / department=按部门 / users=指定用户名列表
type NotificationTarget struct {
	Mode       string   `json:"mode"`
	Department string   `json:"department"`
	Usernames  []string `json:"usernames"`
}

func (s *NotificationService) List(userID uint, page, pageSize int, unreadOnly bool) ([]model.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	tx := s.db.Model(&model.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		tx = tx.Where("is_read = ?", false)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取通知数量失败")
	}
	var items []model.Notification
	if err := tx.Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, errors.New("获取通知列表失败")
	}
	return items, total, nil
}

func (s *NotificationService) UnreadCount(userID uint) (int64, error) {
	var count int64
	if err := s.db.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error; err != nil {
		return 0, errors.New("获取未读数量失败")
	}
	return count, nil
}

func (s *NotificationService) MarkRead(userID, id uint) error {
	res := s.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("通知不存在")
	}
	return nil
}

func (s *NotificationService) MarkAllRead(userID uint) error {
	return s.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": time.Now()}).Error
}

// Send 管理员发送通知：all / department / users 三种目标，逐用户落库并记录发送批次。
func (s *NotificationService) Send(adminID uint, title, content, ntype string, target NotificationTarget) (int, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return 0, errors.New("通知标题不能为空")
	}
	if len([]rune(title)) > 150 {
		return 0, errors.New("通知标题不能超过 150 字")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return 0, errors.New("通知内容不能为空")
	}
	if len([]rune(content)) > 5000 {
		return 0, errors.New("通知内容不能超过 5000 字")
	}
	if ntype == "" {
		ntype = "system"
	}

	var userIDs []uint
	targetDesc := ""
	switch target.Mode {
	case "all":
		targetDesc = "全部会员"
		if err := s.db.Model(&model.User{}).
			Where("role = ? AND status = ?", consts.RoleMember, consts.StatusActive).
			Pluck("id", &userIDs).Error; err != nil {
			return 0, errors.New("查询目标用户失败")
		}
	case "department":
		dep := strings.TrimSpace(target.Department)
		if !consts.IsValidDepartment(dep) {
			return 0, errors.New("请选择正确的部门")
		}
		targetDesc = "部门:" + dep
		if err := s.db.Model(&model.User{}).
			Where("role = ? AND status = ? AND department = ?", consts.RoleMember, consts.StatusActive, dep).
			Pluck("id", &userIDs).Error; err != nil {
			return 0, errors.New("查询目标用户失败")
		}
	case "users":
		names := make([]string, 0, len(target.Usernames))
		for _, n := range target.Usernames {
			if n = strings.TrimSpace(n); n != "" {
				names = append(names, n)
			}
		}
		if len(names) == 0 {
			return 0, errors.New("请填写要发送的用户名")
		}
		if len(names) > 500 {
			return 0, errors.New("单次最多选择 500 个用户")
		}
		targetDesc = fmt.Sprintf("用户:%d 人", len(names))
		if err := s.db.Model(&model.User{}).
			Where("username IN ? AND role = ? AND status = ?", names, consts.RoleMember, consts.StatusActive).
			Pluck("id", &userIDs).Error; err != nil {
			return 0, errors.New("查询目标用户失败")
		}
	default:
		return 0, errors.New("不支持的发送目标")
	}
	if len(userIDs) == 0 {
		return 0, errors.New("没有匹配的目标用户")
	}

	count := 0
	err := s.db.Transaction(func(tx *gorm.DB) error {
		batch := model.NotificationBatch{
			AdminID: adminID,
			Title:   title,
			Content: content,
			Type:    ntype,
			Target:  targetDesc,
			Count:   len(userIDs),
		}
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}
		notices := make([]model.Notification, 0, len(userIDs))
		for _, uid := range userIDs {
			notices = append(notices, model.Notification{
				UserID: uid, Title: title, Content: content, Type: ntype,
			})
		}
		if err := tx.CreateInBatches(notices, 500).Error; err != nil {
			return err
		}
		count = len(userIDs)
		return nil
	})
	if err != nil {
		return 0, errors.New("发送通知失败")
	}
	// 实时推送：在线用户立即收到红点刷新事件
	if s.notifier != nil {
		s.notifier(userIDs)
	}
	return count, nil
}

// Batches 管理员查看发送记录
func (s *NotificationService) Batches(page, pageSize int) ([]model.NotificationBatch, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	var total int64
	if err := s.db.Model(&model.NotificationBatch{}).Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取发送记录失败")
	}
	var items []model.NotificationBatch
	if err := s.db.Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, errors.New("获取发送记录失败")
	}
	return items, total, nil
}
