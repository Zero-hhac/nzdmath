package handler

import (
	"math-top/internal/config"
	"math-top/internal/response"
	"math-top/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

type RegisterRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	RealName   string `json:"real_name" binding:"required"`
	ClassName  string `json:"class_name" binding:"required"`
	Department string `json:"department" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "请完整填写注册信息")
		return
	}
	user, err := h.svc.Register(req.Username, req.Password, req.Nickname, req.Email, req.RealName, req.ClassName, req.Department)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"email":      user.Email,
			"role":       user.Role,
			"real_name":  user.RealName,
			"class_name": user.ClassName,
			"department": user.Department,
		},
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "请输入用户名和密码")
		return
	}
	token, user, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"email":      user.Email,
			"role":       user.Role,
			"avatar":     user.Avatar,
			"real_name":  user.RealName,
			"class_name": user.ClassName,
			"department": user.Department,
		},
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		response.Success(c, nil)
		return
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && parts[0] == "Bearer" {
		_ = h.svc.Logout(parts[1])
	}
	response.Success(c, nil)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

type UserForgotPasswordRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type UserResetPasswordRequest struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ForgotPassword 第一步：校验用户名+邮箱，生成验证码并（尽可能）邮件送达。
// SMTP 未配置时，debug 模式回显验证码便于本地联调；release 模式提示联系管理员。
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req UserForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	code, delivered, err := h.svc.ForgotPassword(req.Username, req.Email)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if delivered {
		response.Success(c, gin.H{"message": "验证码已发送至邮箱，10 分钟内有效"})
		return
	}
	if config.GlobalConfig.App.Mode == "release" {
		response.Fail(c, 500, "邮件服务未配置，请联系管理员重置密码")
		return
	}
	response.Success(c, gin.H{
		"message":  "邮件服务未配置（本地调试模式），验证码：" + code,
		"dev_code": code,
	})
}

// ResetPassword 第二步：验证码 + 新密码完成重置。
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req UserResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.ResetPassword(req.Username, req.Email, req.Code, req.NewPassword); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "密码重置成功，请使用新密码登录"})
}
