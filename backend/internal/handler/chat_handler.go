package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	svc *service.ChatService
}

type SendChatTextRequest struct {
	Content string `json:"content" binding:"required"`
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) Join(c *gin.Context) {
	userID := c.GetUint("user_id")
	count, err := h.svc.Join(userID)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"online_count": count})
}

func (h *ChatHandler) Leave(c *gin.Context) {
	userID := c.GetUint("user_id")
	if err := h.svc.Leave(userID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *ChatHandler) ListMessages(c *gin.Context) {
	userID := c.GetUint("user_id")
	afterID := uint(0)
	if raw := c.Query("after_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Fail(c, 400, "无效的消息 ID")
			return
		}
		afterID = uint(id)
	}
	beforeID := uint(0)
	if raw := c.Query("before_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Fail(c, 400, "无效的历史消息 ID")
			return
		}
		beforeID = uint(id)
	}
	if afterID > 0 && beforeID > 0 {
		response.Fail(c, 400, "after_id 和 before_id 不能同时使用")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	afterDeleteMs, _ := strconv.ParseInt(c.Query("after_delete_ms"), 10, 64)

	result, err := h.svc.ListMessages(userID, afterID, beforeID, limit, afterDeleteMs)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ChatHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的消息 ID")
		return
	}
	if err := h.svc.DeleteMessage(userID, uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *ChatHandler) SendText(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req SendChatTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}

	msg, err := h.svc.SendText(userID, req.Content)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, msg)
}

func (h *ChatHandler) SendFile(c *gin.Context) {
	userID := c.GetUint("user_id")
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "请上传文件")
		return
	}

	msg, err := h.svc.SendFile(userID, file)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, msg)
}
