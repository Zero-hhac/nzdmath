package service

import (
	"errors"
	"math-top/internal/dto"
	"math-top/internal/model"

	"gorm.io/gorm"
)

type AdminEventService struct {
	db *gorm.DB
}

func NewAdminEventService(db *gorm.DB) *AdminEventService {
	return &AdminEventService{db: db}
}

func (s *AdminEventService) List(q dto.AdminListQuery) ([]model.Event, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}

	tx := s.db.Model(&model.Event{})
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.Keyword != "" {
		tx = tx.Where("title LIKE ?", "%"+q.Keyword+"%")
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []model.Event
	if err := tx.Order("id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (s *AdminEventService) Get(id uint) (*model.Event, error) {
	var event model.Event
	if err := s.db.First(&event, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("活动不存在")
		}
		return nil, err
	}
	return &event, nil
}

func (s *AdminEventService) Create(adminID uint, req dto.AdminCreateEventRequest) (*model.Event, error) {
	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("结束时间不能早于开始时间")
	}
	event := model.Event{
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		Category:   req.Category,
		Location:   req.Location,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		CoverUrl:   req.CoverURL,
		Capacity:   req.Capacity,
		Status:     req.Status,
		IsFeatured: req.IsFeatured,
		CreatedBy:  adminID,
	}
	if err := s.db.Create(&event).Error; err != nil {
		return nil, errors.New("创建活动失败")
	}
	return &event, nil
}

func (s *AdminEventService) Update(id uint, req dto.AdminUpdateEventRequest) error {
	if req.EndTime.Before(req.StartTime) {
		return errors.New("结束时间不能早于开始时间")
	}
	res := s.db.Model(&model.Event{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title":       req.Title,
		"summary":     req.Summary,
		"content":     req.Content,
		"category":    req.Category,
		"location":    req.Location,
		"start_time":  req.StartTime,
		"end_time":    req.EndTime,
		"cover_url":   req.CoverURL,
		"capacity":    req.Capacity,
		"status":      req.Status,
		"is_featured": req.IsFeatured,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("活动不存在")
	}
	return nil
}

func (s *AdminEventService) Delete(id uint) error {
	// 级联清理该活动的报名记录，避免孤儿数据
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", id).Delete(&model.EventRegistration{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&model.Event{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("活动不存在")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminEventService) ToggleFeature(id uint) error {
	var event model.Event
	if err := s.db.First(&event, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("活动不存在")
		}
		return err
	}
	return s.db.Model(&event).Update("is_featured", !event.IsFeatured).Error
}

// SetExpired 管理员手动标记活动是否过期：
// 过期活动从公开列表下线、报名被拒绝（我的报名/后台名单仍可见）。
func (s *AdminEventService) SetExpired(id uint, expired bool) error {
	res := s.db.Model(&model.Event{}).Where("id = ?", id).Update("is_expired", expired)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("活动不存在")
	}
	return nil
}
