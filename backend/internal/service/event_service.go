package service

import (
	"errors"
	"math-top/internal/model"

	"gorm.io/gorm"
)

type EventService struct {
	db *gorm.DB
}

func NewEventService(db *gorm.DB) *EventService {
	return &EventService{db: db}
}

func (s *EventService) ListEvents(page, pageSize int) ([]model.Event, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	}
	if pageSize > 100 {
		pageSize = 100
	}
	var total int64
	if err := s.db.Model(&model.Event{}).Where("status = ?", 1).Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取活动数量失败")
	}
	// 过期活动保留在列表中（卡片显示"已过期"），排在未过期之后
	var events []model.Event
	err := s.db.Select("id, title, summary, category, location, start_time, end_time, cover_url, is_featured, is_expired, status, created_at, content").
		Where("status = ?", 1).
		Order("is_expired asc, id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&events).Error
	if err != nil {
		return nil, 0, errors.New("获取活动列表失败")
	}
	return events, total, nil
}

func (s *EventService) GetEventByID(id uint64) (*model.Event, error) {
	var event model.Event
	err := s.db.Where("id = ? AND status = ?", id, 1).First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动不存在")
		}
		return nil, err
	}
	return &event, nil
}