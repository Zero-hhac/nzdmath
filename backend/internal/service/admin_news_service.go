package service

import (
	"errors"
	"math-top/internal/dto"
	"math-top/internal/model"
	"time"

	"gorm.io/gorm"
)

type AdminNewsService struct {
	db *gorm.DB
}

func NewAdminNewsService(db *gorm.DB) *AdminNewsService {
	return &AdminNewsService{db: db}
}

func (s *AdminNewsService) List(q dto.AdminListQuery) ([]model.News, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}

	tx := s.db.Model(&model.News{})
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

	var news []model.News
	if err := tx.Order("id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&news).Error; err != nil {
		return nil, 0, err
	}
	return news, total, nil
}

func (s *AdminNewsService) Get(id uint) (*model.News, error) {
	var news model.News
	if err := s.db.First(&news, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("资讯不存在")
		}
		return nil, err
	}
	return &news, nil
}

func (s *AdminNewsService) Create(adminID uint, req dto.AdminCreateNewsRequest) (*model.News, error) {
	news := model.News{
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		Category:   req.Category,
		Tag:        req.Tag,
		CoverURL:   req.CoverURL,
		Status:     req.Status,
		IsFeatured: req.IsFeatured,
		CreatedBy:  adminID,
	}
	if req.Status == 1 {
		now := time.Now()
		news.PublishedAt = &now
	}
	if err := s.db.Create(&news).Error; err != nil {
		return nil, errors.New("创建资讯失败")
	}
	return &news, nil
}

func (s *AdminNewsService) Update(id uint, req dto.AdminUpdateNewsRequest) error {
	var old model.News
	if err := s.db.First(&old, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("资讯不存在")
		}
		return err
	}

	updates := map[string]interface{}{
		"title":       req.Title,
		"summary":     req.Summary,
		"content":     req.Content,
		"category":    req.Category,
		"tag":         req.Tag,
		"cover_url":   req.CoverURL,
		"status":      req.Status,
		"is_featured": req.IsFeatured,
	}
	// 由草稿变为发布时，自动写入发布时间
	if old.Status == 0 && req.Status == 1 {
		now := time.Now()
		updates["published_at"] = &now
	}
	return s.db.Model(&model.News{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AdminNewsService) Delete(id uint) error {
	res := s.db.Delete(&model.News{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("资讯不存在")
	}
	return nil
}
