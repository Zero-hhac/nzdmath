package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	svc    *service.EventService
	regSvc *service.EventRegistrationService
}

func NewEventHandler(svc *service.EventService, regSvc *service.EventRegistrationService) *EventHandler {
	return &EventHandler{
		svc:    svc,
		regSvc: regSvc,
	}
}

func (h *EventHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	events, total, err := h.svc.ListEvents(page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取活动列表失败")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	}
	response.PageSuccess(c, events, total, page, pageSize)
}

func (h *EventHandler) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	event, err := h.svc.GetEventByID(id)
	if err != nil {
		response.Fail(c, 404, "活动不存在")
		return
	}
	// 可选登录态：带 token 时附加当前用户报名状态
	tokenString := ""
	if parts := strings.Split(c.GetHeader("Authorization"), " "); len(parts) == 2 && parts[0] == "Bearer" {
		tokenString = parts[1]
	}
	response.Success(c, h.regSvc.BuildDetail(event, tokenString))
}
