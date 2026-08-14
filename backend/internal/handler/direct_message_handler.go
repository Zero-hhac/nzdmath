package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DirectMessageHandler struct {
	svc *service.DirectMessageService
}

func NewDirectMessageHandler(svc *service.DirectMessageService) *DirectMessageHandler {
	return &DirectMessageHandler{svc: svc}
}

// GetMyMessages 会员获取私聊历史
func (h *DirectMessageHandler) GetMyMessages(c *gin.Context) {
	userID := c.GetUint("user_id")
	beforeID, _ := strconv.ParseUint(c.Query("before_id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	messages, hasMore, nextBeforeID, err := h.svc.GetMyMessages(userID, uint(beforeID), limit)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{
		"messages":       messages,
		"has_more":       hasMore,
		"next_before_id": nextBeforeID,
	})
}

// SendUserMessage 会员发送文本消息
func (h *DirectMessageHandler) SendUserMessage(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "消息内容不能为空")
		return
	}

	msg, err := h.svc.SendUserMessage(userID, req.Content)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, msg)
}

// SendUserFile 会员发送文件/图片
func (h *DirectMessageHandler) SendUserFile(c *gin.Context) {
	userID := c.GetUint("user_id")

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "请上传文件")
		return
	}

	msg, err := h.svc.SendUserFile(userID, file)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, msg)
}

// MarkUserRead 会员标记已读
func (h *DirectMessageHandler) MarkUserRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if err := h.svc.MarkUserRead(userID); err != nil {
		response.Fail(c, 500, "标记已读失败")
		return
	}
	response.Success(c, nil)
}

// GetUnreadCountForUser 会员获取未读消息数
func (h *DirectMessageHandler) GetUnreadCountForUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	count, err := h.svc.GetUnreadCountForUser(userID)
	if err != nil {
		response.Fail(c, 500, "获取未读数失败")
		return
	}
	response.Success(c, gin.H{"unread_count": count})
}

// AdminListConversations 管理员获取会话列表
func (h *DirectMessageHandler) AdminListConversations(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	conversations, total, err := h.svc.ListConversations(keyword, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取会话列表失败")
		return
	}

	response.Success(c, gin.H{
		"conversations": conversations,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
	})
}

// AdminGetMessages 管理员获取某用户的私聊记录
func (h *DirectMessageHandler) AdminGetMessages(c *gin.Context) {
	targetUID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的用户 ID")
		return
	}

	beforeID, _ := strconv.ParseUint(c.Query("before_id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	messages, hasMore, nextBeforeID, err := h.svc.GetMessagesByUser(uint(targetUID), uint(beforeID), limit)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{
		"messages":       messages,
		"has_more":       hasMore,
		"next_before_id": nextBeforeID,
	})
}

// AdminSendMessage 管理员向用户发送文本消息
func (h *DirectMessageHandler) AdminSendMessage(c *gin.Context) {
	adminID := c.GetUint("admin_id")
	targetUID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的用户 ID")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "回复内容不能为空")
		return
	}

	msg, err := h.svc.SendAdminMessage(adminID, uint(targetUID), req.Content)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, msg)
}

// AdminSendFile 管理员向用户发送文件/图片
func (h *DirectMessageHandler) AdminSendFile(c *gin.Context) {
	adminID := c.GetUint("admin_id")
	targetUID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的用户 ID")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "请上传文件")
		return
	}

	msg, err := h.svc.SendAdminFile(adminID, uint(targetUID), file)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, msg)
}

// AdminMarkRead 管理员标记指定用户的消息为已读
func (h *DirectMessageHandler) AdminMarkRead(c *gin.Context) {
	targetUID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的用户 ID")
		return
	}

	if err := h.svc.MarkAdminRead(uint(targetUID)); err != nil {
		response.Fail(c, 500, "标记已读失败")
		return
	}

	response.Success(c, nil)
}

// AdminGetTotalUnread 管理员获取全局未读私聊数
func (h *DirectMessageHandler) AdminGetTotalUnread(c *gin.Context) {
	count, err := h.svc.GetTotalUnreadForAdmin()
	if err != nil {
		response.Fail(c, 500, "获取未读数失败")
		return
	}
	response.Success(c, gin.H{"unread_count": count})
}
