package service

import (
	"errors"
	"math-top/internal/dto"
	"math-top/internal/model"

	"gorm.io/gorm"
)

type AdminShowcaseService struct {
	db *gorm.DB
}

func NewAdminShowcaseService(db *gorm.DB) *AdminShowcaseService {
	return &AdminShowcaseService{db: db}
}

func (s *AdminShowcaseService) List(q dto.AdminListQuery) ([]model.Showcase, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}

	tx := s.db.Model(&model.Showcase{})
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

	var showcases []model.Showcase
	if err := tx.Order("id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&showcases).Error; err != nil {
		return nil, 0, err
	}
	return showcases, total, nil
}

func (s *AdminShowcaseService) Get(id uint) (*model.Showcase, error) {
	var showcase model.Showcase
	if err := s.db.First(&showcase, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("作品不存在")
		}
		return nil, err
	}
	return &showcase, nil
}

func (s *AdminShowcaseService) Create(req dto.AdminCreateShowcaseRequest) (*model.Showcase, error) {
	showcase := model.Showcase{
		Title:       req.Title,
		Author:      req.Author,
		Field:       req.Field,
		Competition: req.Competition,
		Summary:     req.Summary,
		CoverURL:    req.CoverURL,
		Status:      req.Status,
	}
	if err := s.db.Create(&showcase).Error; err != nil {
		return nil, errors.New("创建作品失败")
	}
	return &showcase, nil
}

func (s *AdminShowcaseService) Update(id uint, req dto.AdminUpdateShowcaseRequest) error {
	res := s.db.Model(&model.Showcase{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title":       req.Title,
		"author":      req.Author,
		"field":       req.Field,
		"competition": req.Competition,
		"summary":     req.Summary,
		"cover_url":   req.CoverURL,
		"status":      req.Status,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("作品不存在")
	}
	return nil
}

func (s *AdminShowcaseService) Delete(id uint) error {
	res := s.db.Delete(&model.Showcase{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("作品不存在")
	}
	return nil
}
