package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type DownloadRepository interface {
	Create(log *model.DownloadLog) error
	ListByUser(userID uint, page, pageSize int) ([]model.DownloadLog, int64, error)
}

type downloadRepository struct{ db *gorm.DB }

func NewDownloadRepository(db *gorm.DB) DownloadRepository {
	return &downloadRepository{db: db}
}

func (r *downloadRepository) Create(log *model.DownloadLog) error {
	return r.db.Create(log).Error
}

func (r *downloadRepository) ListByUser(userID uint, page, pageSize int) ([]model.DownloadLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var total int64
	r.db.Model(&model.DownloadLog{}).Where("user_id = ?", userID).Count(&total)
	var logs []model.DownloadLog
	err := r.db.Where("user_id = ?", userID).
		Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&logs).Error
	return logs, total, err
}