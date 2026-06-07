package service

import (
	"errors"
	"math-top/internal/model"

	"gorm.io/gorm"
)

type NewsService struct {
	db *gorm.DB
}

func NewNewsService(db *gorm.DB) *NewsService {
	return &NewsService{db: db}
}

func (s *NewsService) ListNews() ([]model.News, error) {
	var news []model.News
	err := s.db.Select("id, title, summary, category, tag, cover_url, is_featured, status, published_at, created_at, content").
		Where("status = ?", 1).
		Order("published_at desc, id desc").
		Find(&news).Error
	if err != nil {
		return nil, err
	}
	return news, nil
}

func (s *NewsService) GetNewsByID(id uint64) (*model.News, error) {
	var news model.News
	err := s.db.Where("id = ? AND status = ?", id, 1).First(&news).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("新闻资讯不存在")
		}
		return nil, err
	}
	return &news, nil
}