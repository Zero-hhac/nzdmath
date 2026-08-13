package handler

import (
	"math-top/internal/dto"
	"math-top/internal/response"
	"math-top/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	token, admin, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{
		"token": token,
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
			"nickname": admin.Nickname,
			"role":     admin.Role,
		},
	})
}

func (h *AdminHandler) Logout(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			_ = h.svc.Logout(parts[1])
		}
	}
	response.Success(c, nil)
}

type AdminChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword 管理员自助修改密码（后台“账号设置”）
func (h *AdminHandler) ChangePassword(c *gin.Context) {
	adminID := c.GetUint("user_id")
	var req AdminChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.ChangePassword(adminID, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	data, err := h.svc.GetDashboard()
	if err != nil {
		response.Fail(c, 500, "获取仪表盘数据失败")
		return
	}
	response.Success(c, gin.H{
		"counts":         data.Counts,
		"today_new":      data.TodayNew,
		"trend_7days":    data.Trend,
		"today_activity": data.TodayActivity,
		"activity_trend": data.Activity,
		"total_activity": data.TotalActivity,
		"user_count":     data.Counts.Users,
		"event_count":    data.Counts.Events,
		"news_count":     data.Counts.News,
		"resource_count": data.Counts.Resources,
		"showcase_count": data.Counts.Showcases,
	})
}
