package service

import (
	"errors"
	"math-top/internal/model"

	"gorm.io/gorm"
)

type MemberService struct {
	db *gorm.DB
}

func NewMemberService(db *gorm.DB) *MemberService {
	return &MemberService{
		db: db,
	}
}

func (s *MemberService) GetProfile(userID uint) (*model.User, error) {
	// 获取会员信息
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *MemberService) UpdateProfile(userID uint, nickname string, avatar string, bio string, email string) error {
	// 更新会员信息
	updates := map[string]interface{}{
		"nickname": nickname,
		"avatar":   avatar,
		"bio":      bio,
	}
	if email == "" {
		updates["email"] = nil
	} else {
		updates["email"] = email
	}
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}
