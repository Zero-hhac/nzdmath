package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventRegistrationHandler struct {
	svc *service.EventRegistrationService
}

func NewEventRegistrationHandler(svc *service.EventRegistrationService) *EventRegistrationHandler {
	return &EventRegistrationHandler{svc: svc}
}

// Register 报名活动
func (h *EventRegistrationHandler) Register(c *gin.Context) {
	userID := c.GetUint("user_id")
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	if err := h.svc.Register(uint(eventID), userID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "报名成功"})
}

// Cancel 取消报名
func (h *EventRegistrationHandler) Cancel(c *gin.Context) {
	userID := c.GetUint("user_id")
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	if err := h.svc.Cancel(uint(eventID), userID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "已取消报名"})
}

// MyRegistrations 我的报名
func (h *EventRegistrationHandler) MyRegistrations(c *gin.Context) {
	userID := c.GetUint("user_id")
	items, err := h.svc.MyRegistrations(userID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, items)
}

// AdminList 管理员查看报名名单
func (h *EventRegistrationHandler) AdminList(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	items, err := h.svc.ListByEvent(uint(eventID))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, items)
}

// AdminSummary 后台报名管理页：全部活动 + 报名/签到人数
func (h *EventRegistrationHandler) AdminSummary(c *gin.Context) {
	items, err := h.svc.Summary()
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, items)
}

// AdminCheckin 签到
func (h *EventRegistrationHandler) AdminCheckin(c *gin.Context) {
	eventID, err1 := strconv.ParseUint(c.Param("id"), 10, 64)
	userID, err2 := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err1 != nil || err2 != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.MarkAttended(uint(eventID), uint(userID)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "签到成功"})
}

// AdminCancelCheckin 取消签到
func (h *EventRegistrationHandler) AdminCancelCheckin(c *gin.Context) {
	eventID, err1 := strconv.ParseUint(c.Param("id"), 10, 64)
	userID, err2 := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err1 != nil || err2 != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.CancelCheckin(uint(eventID), uint(userID)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "已取消签到"})
}

// AdminRemove 管理员移除报名
func (h *EventRegistrationHandler) AdminRemove(c *gin.Context) {
	eventID, err1 := strconv.ParseUint(c.Param("id"), 10, 64)
	userID, err2 := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err1 != nil || err2 != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.AdminRemove(uint(eventID), uint(userID)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "已移除报名"})
}
