package service

import (
	"errors"
	"strings"

	"math-top/internal/consts"
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

func (s *MemberService) UpdateProfile(userID uint, nickname string, avatar string, bio string, email string, realName string, className string, department string) error {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	realName = strings.TrimSpace(realName)
	className = strings.TrimSpace(className)
	department = strings.TrimSpace(department)

	// 若未传入必填字段，则继承现有用户资料；若现有资料亦为空，则提示必填
	if realName == "" {
		if user.RealName == "" {
			return errors.New("姓名不能为空")
		}
		realName = user.RealName
	}
	if className == "" {
		if user.ClassName == "" {
			return errors.New("班级不能为空")
		}
		className = user.ClassName
	}
	if department == "" {
		if user.Department == "" {
			return errors.New("请选择正确的部门")
		}
		department = user.Department
	} else if !consts.IsValidDepartment(department) {
		return errors.New("请选择正确的部门")
	}

	if l := len([]rune(nickname)); l > 50 {
		return errors.New("昵称长度不能超过 50 字")
	}
	if l := len([]rune(bio)); l > 2000 {
		return errors.New("个人简介不能超过 2000 字")
	}
	if email != "" {
		email = strings.TrimSpace(email)
		if !emailRegex.MatchString(email) {
			return errors.New("邮箱格式不正确")
		}
		var otherUser model.User
		if err := s.db.Where("email = ? AND id != ?", email, userID).First(&otherUser).Error; err == nil {
			return errors.New("该邮箱已被其他账号使用")
		}
	}
	// 头像必须是本站上传路径或合法 http(s) URL，防止写入任意字符串
	if avatar != "" && len(avatar) <= 500 &&
		!strings.HasPrefix(avatar, "/uploads/avatars/") &&
		!strings.HasPrefix(avatar, "http://") &&
		!strings.HasPrefix(avatar, "https://") {
		return errors.New("头像地址不合法")
	}
	if len(avatar) > 500 {
		return errors.New("头像地址过长")
	}
	if avatar == "" {
		avatar = user.Avatar
	}

	updates := map[string]interface{}{
		"nickname":   nickname,
		"avatar":     avatar,
		"bio":        bio,
		"real_name":  realName,
		"class_name": className,
		"department": department,
	}
	if email == "" {
		updates["email"] = nil
	} else {
		updates["email"] = email
	}
	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return errors.New("更新资料失败")
	}
	return nil
}
