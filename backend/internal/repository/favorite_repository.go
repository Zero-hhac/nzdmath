package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type FavoriteRepository interface {
	Add(userID, targetID uint, targetType string) error
	Remove(userID, favoriteID uint) error
	RemoveByTarget(userID, targetID uint, targetType string) error
	ListByUser(userID uint) ([]model.Favorite, error)
	ListByUserPaged(userID uint, page, pageSize int) ([]model.Favorite, int64, error)
	Exists(userID, targetID uint, targetType string) bool
	CountByTarget(targetType string, targetID uint) int64
}

type favoriteRepository struct{ db *gorm.DB }

func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Add(userID, targetID uint, targetType string) error {
	fav := model.Favorite{
		UserID:     userID,
		TargetID:   targetID,
		TargetType: targetType,
	}
	return r.db.Create(&fav).Error
}

func (r *favoriteRepository) Remove(userID, favoriteID uint) error {
	return r.db.Where("id = ? AND user_id = ?", favoriteID, userID).Delete(&model.Favorite{}).Error
}

func (r *favoriteRepository) RemoveByTarget(userID, targetID uint, targetType string) error {
	return r.db.Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).Delete(&model.Favorite{}).Error
}

func (r *favoriteRepository) ListByUser(userID uint) ([]model.Favorite, error) {
	var favs []model.Favorite
	err := r.db.Where("user_id = ?", userID).Order("id desc").Find(&favs).Error
	return favs, err
}

func (r *favoriteRepository) ListByUserPaged(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var total int64
	r.db.Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&total)
	var favs []model.Favorite
	err := r.db.Where("user_id = ?", userID).Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&favs).Error
	return favs, total, err
}

func (r *favoriteRepository) Exists(userID, targetID uint, targetType string) bool {
	var count int64
	r.db.Model(&model.Favorite{}).Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).Count(&count)
	return count > 0
}

func (r *favoriteRepository) CountByTarget(targetType string, targetID uint) int64 {
	var n int64
	r.db.Model(&model.Favorite{}).Where("target_type = ? AND target_id = ?", targetType, targetID).Count(&n)
	return n
}