package service

import (
	"errors"
	"math-top/internal/model"
	"time"

	"gorm.io/gorm"
)

type FavoriteService struct {
	db *gorm.DB
}

func NewFavoriteService(db *gorm.DB) *FavoriteService {
	return &FavoriteService{
		db: db,
	}
}

func (s *FavoriteService) AddFavorite(userID uint, targetID uint, targetType string) error {
	fav := model.Favorite{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	return s.db.Create(&fav).Error
}

func (s *FavoriteService) RemoveFavorite(favoriteID uint, userID uint) error {
	var fav model.Favorite
	err := s.db.First(&fav, favoriteID).Error
	if err != nil {
		return errors.New("收藏记录不存在")
	}
	if fav.UserID != userID {
		return errors.New("无权删除他人收藏记录")
	}

	return s.db.Delete(&fav).Error
}

type FavoriteItem struct {
	ID            uint      `json:"id"`
	TargetType    string    `json:"target_type"`
	TargetID      uint      `json:"target_id"`
	TargetTitle   string    `json:"target_title"`
	TargetSummary string    `json:"target_summary"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *FavoriteService) ListFavorites(userID uint) ([]FavoriteItem, error) {
	var favs []model.Favorite
	if err := s.db.Where("user_id = ?", userID).Find(&favs).Error; err != nil {
		return nil, err
	}

	var result []FavoriteItem
	for _, fav := range favs {
		item := FavoriteItem{
			ID:         fav.ID,
			TargetType: fav.TargetType,
			TargetID:   fav.TargetID,
			CreatedAt:  fav.CreatedAt,
		}

		switch fav.TargetType {
		case "event":
			var event model.Event
			if err := s.db.Where("id = ?", fav.TargetID).First(&event).Error; err == nil {
				item.TargetTitle = event.Title
				item.TargetSummary = event.Summary
			}

		case "news":
			var news model.News
			if err := s.db.Where("id = ?", fav.TargetID).First(&news).Error; err == nil {
				item.TargetTitle = news.Title
				item.TargetSummary = news.Summary
			}

		case "resource":
			var resource model.Resource
			if err := s.db.Where("id = ?", fav.TargetID).First(&resource).Error; err == nil {
				item.TargetTitle = resource.Title
				item.TargetSummary = resource.Summary
			}

		case "showcase":
			var showcase model.Showcase
			if err := s.db.Where("id = ?", fav.TargetID).First(&showcase).Error; err == nil {
				item.TargetTitle = showcase.Title
				item.TargetSummary = showcase.Summary
			}
		}
		result = append(result, item)
	}
	return result, nil
}
