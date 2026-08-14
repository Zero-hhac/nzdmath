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

var (
	ErrFavoriteNotFound    = errors.New("收藏记录不存在")
	ErrFavoriteForbidden   = errors.New("无权删除他人收藏记录")
	ErrFavoriteInvalidType = errors.New("不支持的收藏类型")
	ErrFavoriteDuplicate   = errors.New("已收藏，请勿重复收藏")
)

var validFavoriteTypes = map[string]bool{
	"event": true, "news": true, "resource": true, "showcase": true,
}

func (s *FavoriteService) validateTarget(targetType string, targetID uint) error {
	var count int64
	var err error
	switch targetType {
	case "event":
		err = s.db.Model(&model.Event{}).Where("id = ? AND status = 1", targetID).Count(&count).Error
	case "news":
		err = s.db.Model(&model.News{}).Where("id = ? AND status = 1", targetID).Count(&count).Error
	case "resource":
		err = s.db.Model(&model.Resource{}).Where("id = ? AND status = 1", targetID).Count(&count).Error
	case "showcase":
		err = s.db.Model(&model.Showcase{}).Where("id = ? AND status = 1", targetID).Count(&count).Error
	default:
		return ErrFavoriteInvalidType
	}
	if err != nil || count == 0 {
		return errors.New("收藏的内容不存在或已下线")
	}
	return nil
}

func (s *FavoriteService) AddFavorite(userID uint, targetID uint, targetType string) error {
	if !validFavoriteTypes[targetType] {
		return ErrFavoriteInvalidType
	}
	// 校验目标主体存在性与公开状态
	if err := s.validateTarget(targetType, targetID); err != nil {
		return err
	}

	var existing model.Favorite
	err := s.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).First(&existing).Error
	if err == nil {
		return ErrFavoriteDuplicate
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	fav := model.Favorite{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	if err := s.db.Create(&fav).Error; err != nil {
		return errors.New("添加收藏失败")
	}
	return nil
}

func (s *FavoriteService) RemoveFavorite(favoriteID uint, userID uint) error {
	var fav model.Favorite
	err := s.db.First(&fav, favoriteID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFavoriteNotFound
		}
		return err
	}
	if fav.UserID != userID {
		return ErrFavoriteForbidden
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
