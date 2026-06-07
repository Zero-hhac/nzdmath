package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type NewsRepository interface {
	ListPublished() ([]model.News, error)
	ListWithFilter(filter NewsFilter) ([]model.News, int64, error)
	GetPublishedByID(id uint) (*model.News, error)
	GetByID(id uint) (*model.News, error)
	Create(news *model.News) error
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	Count() int64
	CountToday() int64
	CountSince(daysAgo int) map[string]int64
}

type NewsFilter struct {
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type newsRepository struct{ db *gorm.DB }

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) ListPublished() ([]model.News, error) {
	var news []model.News
	err := r.db.Select("id, title, summary, category, tag, cover_url, is_featured, status, published_at, created_at").
		Where("status = ?", 1).
		Order("published_at desc, id desc").
		Find(&news).Error
	return news, err
}

func (r *newsRepository) ListWithFilter(f NewsFilter) ([]model.News, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	tx := r.db.Model(&model.News{})
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
	var news []model.News
	if err := tx.Order("id desc").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&news).Error; err != nil {
		return nil, 0, err
	}
	return news, total, nil
}

func (r *newsRepository) GetPublishedByID(id uint) (*model.News, error) {
	var n model.News
	if err := r.db.Where("id = ? AND status = ?", id, 1).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepository) GetByID(id uint) (*model.News, error) {
	var n model.News
	if err := r.db.First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepository) Create(news *model.News) error {
	return r.db.Create(news).Error
}

func (r *newsRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).Updates(fields).Error
}

func (r *newsRepository) Delete(id uint) error {
	return r.db.Delete(&model.News{}, id).Error
}

func (r *newsRepository) Count() int64 {
	var n int64
	r.db.Model(&model.News{}).Count(&n)
	return n
}

func (r *newsRepository) CountToday() int64 {
	var n int64
	r.db.Model(&model.News{}).Where("created_at >= CURDATE()").Count(&n)
	return n
}

func (r *newsRepository) CountSince(daysAgo int) map[string]int64 {
	type Row struct {
		Day   string
		Total int64
	}
	var rows []Row
	since := gorm.Expr("DATE_SUB(CURDATE(), INTERVAL ? DAY)", daysAgo-1)
	r.db.Model(&model.News{}).
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