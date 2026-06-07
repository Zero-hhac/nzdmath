package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetByUsername(username string) (*model.User, error)
	GetByID(id uint) (*model.User, error)
	Create(user *model.User) error
	UpdatePassword(id uint, hashedPassword string) error
	UpdateLastLogin(id uint) error
	UpdateFields(id uint, fields map[string]interface{}) error
	List(filter UserFilter) ([]model.User, int64, error)
	SetStatus(id uint, status int) error
}

type UserFilter struct {
	Page     int
	PageSize int
	Keyword  string
	Status   *int
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("password_hash", hashedPassword).Error
}

func (r *userRepository) UpdateLastLogin(id uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("NOW()")).Error
}

func (r *userRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

func (r *userRepository) List(filter UserFilter) ([]model.User, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	tx := r.db.Model(&model.User{})
	if filter.Status != nil {
		tx = tx.Where("status = ?", *filter.Status)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		tx = tx.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	if err := tx.Order("id desc").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) SetStatus(id uint, status int) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("status", status).Error
}