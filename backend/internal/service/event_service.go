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

func (s *EventService) ListEvents() ([]model.Event, error) {
	var events []model.Event
	err := s.db.Select("id, title, summary, category, location, start_time, end_time, cover_url, is_featured, status, created_at, content").
		Where("status = ?", 1).
		Order("id desc").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
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