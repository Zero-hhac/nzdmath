package service

import (
	"errors"
	"math-top/internal/dto"
	"math-top/internal/model"

	"gorm.io/gorm"
)

type AdminResourceService struct {
	db *gorm.DB
}

func NewAdminResourceService(db *gorm.DB) *AdminResourceService {
	return &AdminResourceService{db: db}
}

func (s *AdminResourceService) List(q dto.AdminListQuery) ([]model.Resource, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}

	tx := s.db.Model(&model.Resource{})
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

	var resources []model.Resource
	if err := tx.Order("id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&resources).Error; err != nil {
		return nil, 0, err
	}
	return resources, total, nil
}

func (s *AdminResourceService) Get(id uint) (*model.Resource, error) {
	var resource model.Resource
	if err := s.db.First(&resource, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("资源不存在")
		}
		return nil, err
	}
	return &resource, nil
}

func (s *AdminResourceService) Update(id uint, req dto.AdminUpdateResourceRequest) error {
	res := s.db.Model(&model.Resource{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title":       req.Title,
		"summary":     req.Summary,
		"category":    req.Category,
		"cover_url":   req.CoverURL,
		"status":      req.Status,
		"is_featured": req.IsFeatured,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("资源不存在")
	}
	return nil
}

func (s *AdminResourceService) Delete(id uint) error {
	res := s.db.Delete(&model.Resource{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("资源不存在")
	}
	return nil
}
