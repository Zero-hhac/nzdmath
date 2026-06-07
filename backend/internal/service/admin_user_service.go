package service

import (
	"errors"
	"math-top/internal/model"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminUserService struct {
	db *gorm.DB
}

func NewAdminUserService(db *gorm.DB) *AdminUserService {
	return &AdminUserService{db: db}
}

func (s *AdminUserService) List(page, pageSize int, keyword string, status *int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	tx := s.db.Model(&model.User{})
	if status != nil {
		tx = tx.Where("status = ?", *status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	if err := tx.Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (s *AdminUserService) SetStatus(id uint, status int) error {
	res := s.db.Model(&model.User{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

func (s *AdminUserService) ResetPassword(id uint, newPassword string) error {
	newPassword = strings.TrimSpace(newPassword)
	if l := len(newPassword); l < 6 || l > 72 {
		return errors.New("新密码长度需在 6-72 之间")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	res := s.db.Model(&model.User{}).Where("id = ?", id).Update("password_hash", string(hashed))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

func (s *AdminUserService) Delete(id uint) error {
	res := s.db.Delete(&model.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

var _ = gorm.ErrRecordNotFound