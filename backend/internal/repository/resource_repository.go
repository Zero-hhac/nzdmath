package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type ResourceRepository interface {
	ListPublished() ([]model.Resource, error)
	ListWithFilter(filter ResourceFilter) ([]model.Resource, int64, error)
	GetPublishedByID(id uint) (*model.Resource, error)
	GetByID(id uint) (*model.Resource, error)
	GetForDownload(id uint) (*model.Resource, error)
	Create(resource *model.Resource) error
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	IncrementView(id uint)
	IncrementDownload(id uint)
	IncrementLike(id uint)
	Count() int64
}

type ResourceFilter struct {
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type resourceRepository struct{ db *gorm.DB }

func NewResourceRepository(db *gorm.DB) ResourceRepository {
	return &resourceRepository{db: db}
}

func (r *resourceRepository) ListPublished() ([]model.Resource, error) {
	var resources []model.Resource
	err := r.db.Select("id, title, summary, category, cover_url, view_count, download_count, like_count, is_featured, file_ext, file_size, created_at").
		Where("status = ?", 1).
		Order("id desc").
		Find(&resources).Error
	return resources, err
}

func (r *resourceRepository) ListWithFilter(f ResourceFilter) ([]model.Resource, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	tx := r.db.Model(&model.Resource{})
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
	var resources []model.Resource
	if err := tx.Order("id desc").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&resources).Error; err != nil {
		return nil, 0, err
	}
	return resources, total, nil
}

func (r *resourceRepository) GetPublishedByID(id uint) (*model.Resource, error) {
	var resource model.Resource
	if err := r.db.Where("id = ? AND status = ?", id, 1).First(&resource).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r *resourceRepository) GetByID(id uint) (*model.Resource, error) {
	var resource model.Resource
	if err := r.db.First(&resource, id).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r *resourceRepository) GetForDownload(id uint) (*model.Resource, error) {
	var resource model.Resource
	if err := r.db.Where("id = ? AND status = ?", id, 1).First(&resource).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r *resourceRepository) Create(resource *model.Resource) error {
	return r.db.Create(resource).Error
}

func (r *resourceRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Resource{}).Where("id = ?", id).Updates(fields).Error
}

func (r *resourceRepository) Delete(id uint) error {
	return r.db.Delete(&model.Resource{}, id).Error
}

func (r *resourceRepository) IncrementView(id uint) {
	r.db.Model(&model.Resource{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))
}

func (r *resourceRepository) IncrementDownload(id uint) {
	r.db.Model(&model.Resource{}).Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + 1"))
}

func (r *resourceRepository) IncrementLike(id uint) {
	r.db.Model(&model.Resource{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))
}

func (r *resourceRepository) Count() int64 {
	var n int64
	r.db.Model(&model.Resource{}).Count(&n)
	return n
}