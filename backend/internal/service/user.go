package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math-top/internal/config"
	"math-top/internal/consts"
	"math-top/internal/middleware"
	"math-top/internal/model"
	"math-top/internal/utils"
	"math/big"
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

func validateRegister(username, password, nickname, email, realName, className, department string) error {
	username = strings.TrimSpace(username)
	if l := len(username); l < 3 || l > 50 {
		return errors.New("用户名长度需在 3-50 之间")
	}
	if err := utils.ValidatePasswordStrength(password); err != nil {
		return err
	}
	if email != "" && !emailRegex.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}
	if l := len(nickname); nickname != "" && l > 50 {
		return errors.New("昵称长度不能超过 50")
	}
	if l := len(realName); l < 1 || l > 50 {
		return errors.New("姓名不能为空且长度不能超过 50")
	}
	if l := len(className); l < 1 || l > 50 {
		return errors.New("班级不能为空且长度不能超过 50")
	}
	if !consts.IsValidDepartment(department) {
		return errors.New("请选择正确的部门")
	}
	return nil
}

func (s *UserService) Register(username, password, nickname, email, realName, className, department string) (*model.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	realName = strings.TrimSpace(realName)
	className = strings.TrimSpace(className)

	if err := validateRegister(username, password, nickname, email, realName, className, department); err != nil {
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
		RealName:     realName,
		ClassName:    className,
		Department:   department,
		Role:         "member",
		Status:       1,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("注册失败")
	}
	return &user, nil
}

func (s *UserService) Login(username, password string) (string, string, bool, *model.User, error) {
	if username == "" || password == "" {
		return "", "", false, nil, errors.New("用户名或密码不能为空")
	}

	user := model.User{}
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		// 用户表未找到，尝试在 admins 表中查找
		var admin model.Admin
		if errAdmin := s.db.Where("username = ?", username).First(&admin).Error; errAdmin == nil {
			if admin.Status == 0 {
				return "", "", false, nil, errors.New("账号已被禁用")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
				return "", "", false, nil, errors.New("用户名或密码错误")
			}
			// 自动创建或同步对应的 user 记录
			user = model.User{
				Username:     admin.Username,
				PasswordHash: admin.PasswordHash,
				Nickname:     admin.Nickname,
				Role:         "admin",
				Status:       1,
				RealName:     "管理员",
				Department:   "管理层",
			}
			s.db.Create(&user)
		} else {
			return "", "", false, nil, errors.New("用户名或密码错误")
		}
	} else {
		if user.Status == 0 {
			return "", "", false, nil, errors.New("账号已被禁用")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			// 密码不匹配时，尝试校验 admins 表中的密码（若为管理员）
			var admin model.Admin
			if errAdmin := s.db.Where("username = ?", username).First(&admin).Error; errAdmin == nil {
				if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) == nil {
					user.PasswordHash = admin.PasswordHash
					user.Role = "admin"
					s.db.Model(&user).Updates(map[string]interface{}{
						"password_hash": admin.PasswordHash,
						"role":          "admin",
					})
				} else {
					return "", "", false, nil, errors.New("用户名或密码错误")
				}
			} else {
				return "", "", false, nil, errors.New("用户名或密码错误")
			}
		}
	}

	isAdmin := user.Role == "admin" || user.Role == "super_admin" || user.Username == "admin"
	if !isAdmin {
		var adminCount int64
		s.db.Model(&model.Admin{}).Where("username = ?", user.Username).Count(&adminCount)
		if adminCount > 0 {
			isAdmin = true
			user.Role = "admin"
			s.db.Model(&user).Update("role", "admin")
		}
	}

	userToken, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", "", false, nil, errors.New("生成Token失败")
	}

	ttl := time.Duration(config.GlobalConfig.JWT.ExpireHours) * time.Hour
	redisKey := middleware.UserTokenPrefix + userToken
	if err := s.rdb.Set(context.Background(), redisKey, user.ID, ttl).Err(); err != nil {
		return "", "", false, nil, errors.New("redis保存token失败")
	}

	var adminToken string
	if isAdmin {
		adminToken, err = utils.GenerateToken(user.ID, user.Username, "admin")
		if err == nil {
			adminRedisKey := middleware.AdminTokenPrefix + adminToken
			_ = s.rdb.Set(context.Background(), adminRedisKey, user.ID, 24*time.Hour).Err()
		}
	}

	now := time.Now()
	s.db.Model(&user).Update("last_login_at", &now)

	return userToken, adminToken, isAdmin, &user, nil
}

func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
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
	if err := s.db.Model(&user).Update("password_hash", string(hashed)).Error; err != nil {
		return errors.New("修改密码失败")
	}
	// 同步更新管理员表
	if user.Role == "admin" || user.Role == "super_admin" || user.Username == "admin" {
		s.db.Model(&model.Admin{}).Where("username = ?", user.Username).Update("password_hash", string(hashed))
	}
	return nil
}

func (s *UserService) Logout(token string) error {
	redisKey := middleware.UserTokenPrefix + token
	return s.rdb.Del(context.Background(), redisKey).Err()
}

// ForgotPassword 校验用户名+邮箱并生成 6 位重置验证码（10 分钟有效）。
// delivered 表示验证码是否已通过邮件送达；SMTP 未配置/发送失败时为 false，
// 由 handler 决定是否在 debug 模式回显验证码（本地联调用）。
func (s *UserService) ForgotPassword(username, email string) (code string, delivered bool, err error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if username == "" || email == "" {
		return "", false, errors.New("请填写用户名和邮箱")
	}
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		// 统一提示，避免账号枚举
		return "", false, errors.New("用户名或邮箱不匹配")
	}
	if user.Email == nil || *user.Email != email {
		return "", false, errors.New("用户名或邮箱不匹配")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", false, errors.New("系统繁忙，请稍后再试")
	}
	code = fmt.Sprintf("%06d", n.Int64())
	key := "pwdreset:" + username
	if err := s.rdb.Set(context.Background(), key, code, 10*time.Minute).Err(); err != nil {
		return "", false, errors.New("系统繁忙，请稍后再试")
	}
	s.rdb.Del(context.Background(), "pwdreset_attempts:"+username)

	if err := SendPasswordResetCode(email, username, code); err != nil {
		slog.Warn("找回密码邮件发送失败", "username", username, "err", err)
		return code, false, nil
	}
	return "", true, nil
}

// ResetPassword 用验证码重置密码；错误尝试 5 次后验证码作废。
func (s *UserService) ResetPassword(username, email, code, newPassword string) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	code = strings.TrimSpace(code)
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("用户名或邮箱不匹配")
	}
	if user.Email == nil || *user.Email != email {
		return errors.New("用户名或邮箱不匹配")
	}

	key := "pwdreset:" + username
	attemptKey := "pwdreset_attempts:" + username
	stored, err := s.rdb.Get(context.Background(), key).Result()
	if err != nil || stored != code {
		attempts, _ := s.rdb.Incr(context.Background(), attemptKey).Result()
		if attempts >= 5 {
			s.rdb.Del(context.Background(), key, attemptKey)
		} else {
			s.rdb.Expire(context.Background(), attemptKey, 10*time.Minute)
		}
		return errors.New("验证码错误或已过期")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	if err := s.db.Model(&user).Update("password_hash", string(hashed)).Error; err != nil {
		return errors.New("重置密码失败")
	}
	// 同步更新管理员表
	if user.Role == "admin" || user.Role == "super_admin" || user.Username == "admin" {
		s.db.Model(&model.Admin{}).Where("username = ?", user.Username).Update("password_hash", string(hashed))
	}
	s.rdb.Del(context.Background(), key, attemptKey)
	return nil
}
