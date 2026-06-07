package service

import (
	"context"
	"errors"
	"math-top/internal/config"
	"math-top/internal/middleware"
	"math-top/internal/model"
	"math-top/internal/utils"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewUserService(db *gorm.DB, rdb *redis.Client) *UserService {
	return &UserService{
		db:  db,
		rdb: rdb,
	}
}

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func validateRegister(username, password, nickname, email string) error {
	username = strings.TrimSpace(username)
	if l := len(username); l < 3 || l > 50 {
		return errors.New("用户名长度需在 3-50 之间")
	}
	if l := len(password); l < 6 || l > 72 {
		return errors.New("密码长度需在 6-72 之间")
	}
	if email != "" && !emailRegex.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}
	if l := len(nickname); nickname != "" && l > 50 {
		return errors.New("昵称长度不能超过 50")
	}
	return nil
}

func (s *UserService) Register(username, password, nickname, email string) (*model.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if err := validateRegister(username, password, nickname, email); err != nil {
		return nil, err
	}

	var existing model.User
	if err := s.db.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}
	if email != "" {
		var emailExists model.User
		if err := s.db.Where("email = ?", email).First(&emailExists).Error; err == nil {
			return nil, errors.New("邮箱已被注册")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	if nickname == "" {
		nickname = username
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	user := model.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Nickname:     nickname,
		Email:        emailPtr,
		Role:         "member",
		Status:       1,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("注册失败")
	}
	return &user, nil
}

func (s *UserService) Login(username, password string) (string, *model.User, error) {
	if username == "" || password == "" {
		return "", nil, errors.New("用户名或密码不能为空")
	}

	user := model.User{}
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	if user.Status == 0 {
		return "", nil, errors.New("账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", nil, errors.New("生成Token失败")
	}

	ttl := time.Duration(config.GlobalConfig.JWT.ExpireHours) * time.Hour
	redisKey := middleware.UserTokenPrefix + token
	if err := s.rdb.Set(context.Background(), redisKey, user.ID, ttl).Err(); err != nil {
		return "", nil, errors.New("redis保存token失败")
	}

	now := time.Now()
	s.db.Model(&user).Update("last_login_at", &now)

	return token, &user, nil
}

func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	if l := len(newPassword); l < 6 || l > 72 {
		return errors.New("新密码长度需在 6-72 之间")
	}
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	return s.db.Model(&user).Update("password_hash", string(hashed)).Error
}

func (s *UserService) Logout(token string) error {
	redisKey := middleware.UserTokenPrefix + token
	return s.rdb.Del(context.Background(), redisKey).Err()
}