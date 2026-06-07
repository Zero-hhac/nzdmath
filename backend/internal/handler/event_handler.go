package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{
		svc: svc,
	}
}

func (h *EventHandler) List(c *gin.Context) {
	events, err := h.svc.ListEvents()
	if err != nil {
		response.Fail(c, 500, "获取活动列表失败")
		return
	}
	response.Success(c, events)
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
	response.Success(c, event)
}
