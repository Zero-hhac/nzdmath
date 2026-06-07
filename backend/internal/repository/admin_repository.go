package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type AdminRepository interface {
	GetByUsername(username string) (*model.Admin, error)
	Count() int64
	Create(admin *model.Admin) error
	UpdateLastLogin(id uint) error
}

type adminRepository struct{ db *gorm.DB }

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetByUsername(username string) (*model.Admin, error) {
	var a model.Admin
	if err := r.db.Where("username = ?", username).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *adminRepository) Count() int64 {
	var n int64
	r.db.Model(&model.Admin{}).Count(&n)
	return n
}

func (r *adminRepository) Create(admin *model.Admin) error {
	return r.db.Create(admin).Error
}

func (r *adminRepository) UpdateLastLogin(id uint) error {
	return r.db.Model(&model.Admin{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("NOW()")).Error
}