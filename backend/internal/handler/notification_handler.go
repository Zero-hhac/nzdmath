package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// List 会员通知列表
func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	unreadOnly := c.DefaultQuery("unread_only", "0") == "1"
	items, total, err := h.svc.List(userID, page, pageSize, unreadOnly)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	response.PageSuccess(c, items, total, page, pageSize)
}

// UnreadCount 未读数量（导航角标用）
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.GetUint("user_id")
	count, err := h.svc.UnreadCount(userID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, gin.H{"count": count})
}

// MarkRead 标记单条已读
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的通知 ID")
		return
	}
	if err := h.svc.MarkRead(userID, uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// MarkAllRead 全部已读
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if err := h.svc.MarkAllRead(userID); err != nil {
		response.Fail(c, 500, "操作失败")
		return
	}
	response.Success(c, nil)
}

// Send 管理员发送通知（单个/部门/全部）
func (h *NotificationHandler) Send(c *gin.Context) {
	adminID := c.GetUint("user_id")
	var req struct {
		Title   string                     `json:"title" binding:"required"`
		Content string                     `json:"content" binding:"required"`
		Type    string                     `json:"type"`
		Target  service.NotificationTarget `json:"target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	count, err := h.svc.Send(adminID, req.Title, req.Content, req.Type, req.Target)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"count": count, "message": "通知已发送"})
}

// Batches 管理员发送记录
func (h *NotificationHandler) Batches(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, total, err := h.svc.Batches(page, pageSize)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	response.PageSuccess(c, items, total, page, pageSize)
}
