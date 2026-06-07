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
		"user_count":     data.Counts.Users,
		"event_count":    data.Counts.Events,
		"news_count":     data.Counts.News,
		"resource_count": data.Counts.Resources,
		"showcase_count": data.Counts.Showcases,
	})
}
