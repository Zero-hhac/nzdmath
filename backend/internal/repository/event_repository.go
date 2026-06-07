package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type EventRepository interface {
	ListPublished() ([]model.Event, error)
	ListWithFilter(filter EventFilter) ([]model.Event, int64, error)
	GetPublishedByID(id uint) (*model.Event, error)
	GetByID(id uint) (*model.Event, error)
	Create(event *model.Event) error
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	ToggleFeatured(id uint) error
	Count() int64
	CountToday() int64
	CountSince(daysAgo int) map[string]int64
}

type EventFilter struct {
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type eventRepository struct{ db *gorm.DB }

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) ListPublished() ([]model.Event, error) {
	var events []model.Event
	err := r.db.Select("id, title, summary, category, location, start_time, end_time, cover_url, is_featured, status, created_at").
		Where("status = ?", 1).
		Order("id desc").
		Find(&events).Error
	return events, err
}

func (r *eventRepository) ListWithFilter(f EventFilter) ([]model.Event, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	tx := r.db.Model(&model.Event{})
	if f.Status != nil {
		tx = tx.Where("status = ?", *f.Status)
	}
	if f.Keyword != "" {
		tx = tx.Where("title LIKE ?", "%"+f.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var events []model.Event
	if err := tx.Order("id desc").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *eventRepository) GetPublishedByID(id uint) (*model.Event, error) {
	var e model.Event
	if err := r.db.Where("id = ? AND status = ?", id, 1).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *eventRepository) GetByID(id uint) (*model.Event, error) {
	var e model.Event
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *eventRepository) Create(event *model.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Event{}).Where("id = ?", id).Updates(fields).Error
}

func (r *eventRepository) Delete(id uint) error {
	return r.db.Delete(&model.Event{}, id).Error
}

func (r *eventRepository) ToggleFeatured(id uint) error {
	var e model.Event
	if err := r.db.First(&e, id).Error; err != nil {
		return err
	}
	return r.db.Model(&e).Update("is_featured", !e.IsFeatured).Error
}

func (r *eventRepository) Count() int64 {
	var n int64
	r.db.Model(&model.Event{}).Count(&n)
	return n
}

func (r *eventRepository) CountToday() int64 {
	var n int64
	r.db.Model(&model.Event{}).Where("created_at >= CURDATE()").Count(&n)
	return n
}

func (r *eventRepository) CountSince(daysAgo int) map[string]int64 {
	type Row struct {
		Day   string
		Total int64
	}
	var rows []Row
	since := gorm.Expr("DATE_SUB(CURDATE(), INTERVAL ? DAY)", daysAgo-1)
	r.db.Model(&model.Event{}).
		Select("DATE(created_at) as day, COUNT(*) as total").
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Scan(&rows)
	result := make(map[string]int64)
	for _, row := range rows {
		result[row.Day] = row.Total
	}
	return result
}