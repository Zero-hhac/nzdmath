package service

import (
	"errors"
	"math-top/internal/consts"
	"math-top/internal/model"
	"math-top/internal/utils"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminUserService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewAdminUserService(db *gorm.DB, rdb *redis.Client) *AdminUserService {
	return &AdminUserService{db: db, rdb: rdb}
}

// assertNotAdminRow 单用户管理接口的越权防护（#6）：
// 会员表中角色为管理员/超级管理员或用户名为 admin 的行，一律拒绝普通管理员操作，
// 与批量接口（role NOT IN）的防护对齐，防止普通 admin 重置 super_admin 密码。
func (s *AdminUserService) assertNotAdminRow(user model.User) error {
	if user.Role == consts.RoleAdmin || user.Role == consts.RoleSuperAdmin || strings.EqualFold(user.Username, "admin") {
		return errors.New("管理员账号请在管理员账号体系中自助管理")
	}
	return nil
}

func (s *AdminUserService) List(page, pageSize int, keyword string, status *int, department string, incomplete bool) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	tx := s.db.Model(&model.User{})
	if status != nil {
		tx = tx.Where("status = ?", *status)
	}
	if department != "" {
		tx = tx.Where("department = ?", department)
	}
	if incomplete {
		tx = tx.Where("(real_name = '' OR class_name = '' OR department = '')")
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ? OR real_name LIKE ?", like, like, like, like)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	if err := tx.Order("CASE WHEN real_name = '' OR real_name IS NULL THEN 1 ELSE 0 END, real_name asc, id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (s *AdminUserService) SetStatus(id uint, status int) error {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	// #6：管理员行不参与启停
	if err := s.assertNotAdminRow(user); err != nil {
		return err
	}
	res := s.db.Model(&model.User{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	// #5：封禁时立即吊销该用户全部会话
	if status == consts.StatusDisabled {
		RevokeTokensForUserID(s.rdb, s.db, id)
	}
	return nil
}

func (s *AdminUserService) ResetPassword(id uint, newPassword string) error {
	newPassword = strings.TrimSpace(newPassword)
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	// #6：管理员行不参与重置密码
	if err := s.assertNotAdminRow(user); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	res := s.db.Model(&user).Update("password_hash", string(hashed))
	if res.Error != nil {
		return res.Error
	}
	// #5：重置密码后吊销该用户全部会话
	RevokeTokensForUserID(s.rdb, s.db, id)
	return nil
}

func (s *AdminUserService) Delete(id uint) error {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	// #6：管理员行不参与删除
	if err := s.assertNotAdminRow(user); err != nil {
		return err
	}
	res := s.db.Delete(&model.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	// #5：删除用户后吊销其全部会话
	RevokeTokensForUserID(s.rdb, s.db, id)
	return nil
}

// BatchSetStatus 批量启/停用。管理员账号（admin/super_admin）不参与，防止把后台锁死。
func (s *AdminUserService) BatchSetStatus(ids []uint, status int) (int, error) {
	if len(ids) == 0 {
		return 0, errors.New("请选择至少一个用户")
	}
	res := s.db.Model(&model.User{}).
		Where("id IN ? AND role NOT IN ?", ids, []string{consts.RoleAdmin, consts.RoleSuperAdmin}).
		Update("status", status)
	if res.Error != nil {
		return 0, res.Error
	}
	// #5：批量封禁同样吊销会话
	if status == consts.StatusDisabled && res.RowsAffected > 0 {
		for _, id := range ids {
			RevokeTokensForUserID(s.rdb, s.db, id)
		}
	}
	return int(res.RowsAffected), nil
}

// BatchResetPassword 批量重置密码：bcrypt 只 hash 一次，批量应用到选中用户。
// 管理员账号同样排除，避免改动管理端凭据。
func (s *AdminUserService) BatchResetPassword(ids []uint, newPassword string) (int, error) {
	if len(ids) == 0 {
		return 0, errors.New("请选择至少一个用户")
	}
	newPassword = strings.TrimSpace(newPassword)
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return 0, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, errors.New("密码加密失败")
	}
	res := s.db.Model(&model.User{}).
		Where("id IN ? AND role NOT IN ?", ids, []string{consts.RoleAdmin, consts.RoleSuperAdmin}).
		Update("password_hash", string(hashed))
	if res.Error != nil {
		return 0, res.Error
	}
	// #5：批量重置后吊销会话
	if res.RowsAffected > 0 {
		for _, id := range ids {
			RevokeTokensForUserID(s.rdb, s.db, id)
		}
	}
	return int(res.RowsAffected), nil
}

// BatchDelete 事务批量软删用户（gorm Delete 走 DeletedAt）。管理员账号不参与。
func (s *AdminUserService) BatchDelete(ids []uint) (int, error) {
	if len(ids) == 0 {
		return 0, errors.New("请选择至少一个用户")
	}
	var affected int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("id IN ? AND role NOT IN ?", ids, []string{consts.RoleAdmin, consts.RoleSuperAdmin}).
			Delete(&model.User{})
		if res.Error != nil {
			return res.Error
		}
		affected = res.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	// #5：删除后吊销会话
	if affected > 0 {
		for _, id := range ids {
			RevokeTokensForUserID(s.rdb, s.db, id)
		}
	}
	return int(affected), nil
}

func (s *AdminUserService) ExportExcel(department string) (*excelize.File, error) {
	tx := s.db.Model(&model.User{})
	if department != "" {
		tx = tx.Where("department = ?", department)
	}
	var users []model.User
	if err := tx.Order("CASE WHEN real_name = '' OR real_name IS NULL THEN 1 ELSE 0 END, real_name asc, id desc").
		Find(&users).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	const sheet = "会员名单"
	_ = f.SetSheetName("Sheet1", sheet)

	headers := []string{"姓名", "班级", "部门"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for i, u := range users {
		realName := u.RealName
		if realName == "" {
			realName = "未填写"
		}
		className := u.ClassName
		if className == "" {
			className = "未填写"
		}
		dep := u.Department
		if dep == "" {
			dep = "未分配"
		}
		row := i + 2
		c1, _ := excelize.CoordinatesToCellName(1, row)
		c2, _ := excelize.CoordinatesToCellName(2, row)
		c3, _ := excelize.CoordinatesToCellName(3, row)
		_ = f.SetCellValue(sheet, c1, realName)
		_ = f.SetCellValue(sheet, c2, className)
		_ = f.SetCellValue(sheet, c3, dep)
	}
	return f, nil
}
