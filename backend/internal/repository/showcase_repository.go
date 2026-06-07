package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type ShowcaseRepository interface {
	ListPublishedWithFilter(field, competition, keyword string, page, pageSize int) ([]model.Showcase, int64, error)
	ListWithFilter(filter ShowcaseFilter) ([]model.Showcase, int64, error)
	GetPublishedByID(id uint) (*model.Showcase, error)
	GetByID(id uint) (*model.Showcase, error)
	Create(showcase *model.Showcase) error
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	IncrementView(id uint)
	Count() int64
}

type ShowcaseFilter struct {
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type showcaseRepository struct{ db *gorm.DB }

func NewShowcaseRepository(db *gorm.DB) ShowcaseRepository {
	return &showcaseRepository{db: db}
}

func (r *showcaseRepository) ListPublishedWithFilter(field, competition, keyword string, page, pageSize int) ([]model.Showcase, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	tx := r.db.Model(&model.Showcase{}).Where("status = ?", 1)
	if field != "" && field != "全部领域" {
		tx = tx.Where("field = ?", field)
	}
	if competition != "" && competition != "全部赛事" {
		tx = tx.Where("competition = ?", competition)
	}
	if keyword != "" {
		tx = tx.Where("title LIKE ? OR author LIKE ? OR summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var showcases []model.Showcase
	if err := tx.Order("view_count desc, id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&showcases).Error; err != nil {
		return nil, 0, err
	}
	return showcases, total, nil
}

func (r *showcaseRepository) ListWithFilter(f ShowcaseFilter) ([]model.Showcase, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	tx := r.db.Model(&model.Showcase{})
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
	var showcases []model.Showcase
	if err := tx.Order("id desc").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&showcases).Error; err != nil {
		return nil, 0, err
	}
	return showcases, total, nil
}

func (r *showcaseRepository) GetPublishedByID(id uint) (*model.Showcase, error) {
	var s model.Showcase
	if err := r.db.Where("id = ? AND status = ?", id, 1).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *showcaseRepository) GetByID(id uint) (*model.Showcase, error) {
	var s model.Showcase
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *showcaseRepository) Create(showcase *model.Showcase) error {
	return r.db.Create(showcase).Error
}

func (r *showcaseRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Showcase{}).Where("id = ?", id).Updates(fields).Error
}

func (r *showcaseRepository) Delete(id uint) error {
	return r.db.Delete(&model.Showcase{}, id).Error
}

func (r *showcaseRepository) IncrementView(id uint) {
	r.db.Model(&model.Showcase{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))
}

func (r *showcaseRepository) Count() int64 {
	var n int64
	r.db.Model(&model.Showcase{}).Count(&n)
	return n
}