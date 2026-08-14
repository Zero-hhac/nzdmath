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

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+.[A-Za-z]{2,}$`)

// reservedUsernames 保留用户名黑名单：防止管理员账号创建前被抢注（#1）。
var reservedUsernames = map[string]bool{
	"admin": true, "administrator": true, "root": true, "system": true,
	"superadmin": true, "super_admin": true, "sysadmin": true, "support": true,
	"test": true, "guest": true, "operator": true, "service": true,
}

func isReservedUsername(username string) bool {
	return reservedUsernames[strings.ToLower(strings.TrimSpace(username))]
}

func validateRegister(username, password, nickname, email, realName, className, department string) error {
	username = strings.TrimSpace(username)
	if l := len(username); l < 3 || l > 50 {
		return errors.New("用户名长度需在 3-50 之间")
	}
	if err := utils.ValidatePasswordStrength(password); err != nil {
		return err
	}
	if email == "" || !emailRegex.MatchString(email) {
		return errors.New("请输入有效的邮箱地址")
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

// SendRegisterCode 校验邮箱未注册并生成 6 位注册验证码（10 分钟有效，60 秒防刷冷却）。
// 已注册邮箱与成功发出返回同一提示（后台静默不发），防止邮箱枚举（#15）。
func (s *UserService) SendRegisterCode(email string) (code string, delivered bool, err error) {
	email = strings.TrimSpace(email)
	if email == "" || !emailRegex.MatchString(email) {
		return "", false, errors.New("请输入有效的邮箱地址")
	}

	// 检查邮箱是否已被注册：已注册时静默返回“已发送”，不暴露枚举信息
	var existing model.User
	if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
		return "", true, nil
	}

	// 检查 60 秒防刷冷却
	cooldownKey := "regcode_cooldown:" + email
	if exists, _ := s.rdb.Exists(context.Background(), cooldownKey).Result(); exists > 0 {
		return "", false, errors.New("验证码发送太频繁，请稍后再试")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", false, errors.New("系统繁忙，请稍后再试")
	}
	code = fmt.Sprintf("%06d", n.Int64())
	key := "regcode:" + email
	if err := s.rdb.Set(context.Background(), key, code, 10*time.Minute).Err(); err != nil {
		return "", false, errors.New("系统繁忙，请稍后再试")
	}
	s.rdb.Set(context.Background(), cooldownKey, "1", 60*time.Second)

	if err := SendRegisterVerifyCode(email, code); err != nil {
		slog.Warn("注册验证码邮件发送失败", "email", email, "err", err)
		return code, false, nil
	}
	return "", true, nil
}

func (s *UserService) Register(username, password, nickname, email, code, realName, className, department string) (*model.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	code = strings.TrimSpace(code)
	realName = strings.TrimSpace(realName)
	className = strings.TrimSpace(className)

	if err := validateRegister(username, password, nickname, email, realName, className, department); err != nil {
		return nil, err
	}
	if code == "" {
		return nil, errors.New("请输入邮箱验证码")
	}

	// 校验验证码（#7：错误 5 次后作废，要求重新获取）
	regKey := "regcode:" + email
	attemptKey := "regcode_attempts:" + email
	savedCode, err := s.rdb.Get(context.Background(), regKey).Result()
	if err == redis.Nil {
		return nil, errors.New("验证码已过期或不存在，请重新获取")
	} else if err != nil {
		return nil, errors.New("系统繁忙，请稍后再试")
	}
	if savedCode != code {
		attempts, _ := s.rdb.Incr(context.Background(), attemptKey).Result()
		if attempts >= 5 {
			s.rdb.Del(context.Background(), regKey, attemptKey)
			return nil, errors.New("验证码错误次数过多，请重新获取")
		}
		s.rdb.Expire(context.Background(), attemptKey, 10*time.Minute)
		return nil, errors.New("验证码错误")
	}
	s.rdb.Del(context.Background(), attemptKey)

	// 用户名查重：会员表 + 管理员表 + 保留名黑名单（#1）
	if isReservedUsername(username) {
		return nil, errors.New("用户名已存在")
	}
	var existing model.User
	if err := s.db.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}
	var adminExists model.Admin
	if err := s.db.Where("username = ?", username).First(&adminExists).Error; err == nil {
		// 与“用户名已存在”提示保持一致，防止探测管理员账号
		return nil, errors.New("用户名已存在")
	}
	var emailExists model.User
	if err := s.db.Where("email = ?", email).First(&emailExists).Error; err == nil {
		return nil, errors.New("邮箱已被注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	if nickname == "" {
		nickname = username
	}

	user := model.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Nickname:     nickname,
		Email:        &email,
		RealName:     realName,
		ClassName:    className,
		Department:   department,
		Role:         "member",
		Status:       1,
	}

	// #28：不向外透传数据库错误细节，仅记服务端日志
	if err := s.db.Create(&user).Error; err != nil {
		slog.Error("注册创建用户失败", "username", username, "err", err)
		return nil, errors.New("注册失败，请稍后再试")
	}

	// 注册成功后清除验证码
	s.rdb.Del(context.Background(), regKey)

	return &user, nil
}

// 登录时序侧信道抹平：用户不存在时也执行一次等代价的假密码比较（#15）。
// 使用真实 bcrypt 哈希（"password" 的哈希），确保比较耗时与真实校验一致。
var dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

func fakePasswordCompare() {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte("invalid-password-for-timing"))
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
				// #15：禁用提示并入“用户名或密码错误”，避免枚举
				fakePasswordCompare()
				return "", "", false, nil, errors.New("用户名或密码错误")
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
			// #25：创建失败时拒绝签发 token，避免产生 user_id=0 的合法 token
			if err := s.db.Create(&user).Error; err != nil {
				slog.Error("管理员登录自动建会员记录失败", "username", username, "err", err)
				return "", "", false, nil, errors.New("登录失败，请稍后再试")
			}
		} else {
			// #15：抹平用户不存在的响应时间差异
			fakePasswordCompare()
			return "", "", false, nil, errors.New("用户名或密码错误")
		}
	} else {
		if user.Status == 0 {
			// #15：禁用提示并入“用户名或密码错误”
			fakePasswordCompare()
			return "", "", false, nil, errors.New("用户名或密码错误")
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

	// #1：管理员身份只由角色字段与管理员表实际记录决定，不再特判用户名 “admin”
	isAdmin := user.Role == "admin" || user.Role == "super_admin"
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
	// #5：登记用户→token 索引，改密/封禁时可统一吊销
	IndexUserToken(s.rdb, user.ID, userToken, middleware.UserTokenPrefix, ttl)

	var adminToken string
	if isAdmin {
		adminToken, err = utils.GenerateToken(user.ID, user.Username, "admin")
		if err == nil {
			adminRedisKey := middleware.AdminTokenPrefix + adminToken
			// #5：admin_token 有效期读取配置，不再硬编码 24 小时
			if err := s.rdb.Set(context.Background(), adminRedisKey, user.ID, ttl).Err(); err == nil {
				IndexUserToken(s.rdb, user.ID, adminToken, middleware.AdminTokenPrefix, ttl)
			}
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
	if user.Role == "admin" || user.Role == "super_admin" {
		s.db.Model(&model.Admin{}).Where("username = ?", user.Username).Update("password_hash", string(hashed))
	}
	// #5：改密后吊销该用户全部既有 token
	RevokeTokensForUserID(s.rdb, s.db, userID)
	return nil
}

func (s *UserService) Logout(token string) error {
	redisKey := middleware.UserTokenPrefix + token
	if err := s.rdb.Del(context.Background(), redisKey).Err(); err != nil {
		return err
	}
	// #5：登出时同步从用户索引移除
	if claims, err := utils.ParseToken(token); err == nil {
		UnindexUserToken(s.rdb, claims.UserID, token, middleware.UserTokenPrefix)
	}
	return nil
}

// ForgotPassword 校验用户名+邮箱并生成 6 位重置验证码（10 分钟有效）。
// #8：按“用户名+邮箱”维度 60 秒发送冷却，防邮件轰炸；尝试计数只在
// 新验证码生成时重置（受冷却限制），普通发码请求不再无条件清零计数。
// delivered 表示验证码是否已通过邮件送达；SMTP 未配置/发送失败时为 false，
// 由 handler 决定是否在 debug 模式回显验证码（本地联调用）。
func (s *UserService) ForgotPassword(username, email string) (code string, delivered bool, err error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if username == "" || email == "" {
		return "", false, errors.New("请填写用户名和邮箱")
	}

	// 发送冷却（#8）
	cooldownKey := "pwdreset_cooldown:" + strings.ToLower(username) + ":" + strings.ToLower(email)
	if exists, _ := s.rdb.Exists(context.Background(), cooldownKey).Result(); exists > 0 {
		return "", false, errors.New("验证码发送太频繁，请稍后再试")
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
	s.rdb.Set(context.Background(), cooldownKey, "1", 60*time.Second)
	// 新验证码生成时才重置尝试计数（#8）
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
	if user.Role == "admin" || user.Role == "super_admin" {
		s.db.Model(&model.Admin{}).Where("username = ?", user.Username).Update("password_hash", string(hashed))
	}
	s.rdb.Del(context.Background(), key, attemptKey)
	// #5：重置密码后吊销该用户全部既有 token
	RevokeTokensForUserID(s.rdb, s.db, user.ID)
	return nil
}

// CleanupAdminShadowAccounts 数据清洗（#1）：修复前被抢注的管理员同名会员记录，
// 将密码哈希与角色同步为管理员表值，消除提权隐患。启动时调用。
func (s *UserService) CleanupAdminShadowAccounts() error {
	var admins []model.Admin
	if err := s.db.Select("username, password_hash, role").Find(&admins).Error; err != nil {
		return err
	}
	for _, a := range admins {
		res := s.db.Model(&model.User{}).
			Where("username = ? AND role != ?", a.Username, a.Role).
			Updates(map[string]interface{}{
				"password_hash": a.PasswordHash,
				"role":          a.Role,
			})
		if res.Error != nil {
			slog.Warn("管理员同名会员记录清洗失败", "username", a.Username, "err", res.Error)
		} else if res.RowsAffected > 0 {
			slog.Warn("已清洗管理员同名会员记录", "username", a.Username, "rows", res.RowsAffected)
		}
	}
	return nil
}
