package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminChatHandler struct {
	svc *service.ChatService
}

func NewAdminChatHandler(svc *service.ChatService) *AdminChatHandler {
	return &AdminChatHandler{svc: svc}
}

func (h *AdminChatHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	messages, total, err := h.svc.AdminListMessages(page, pageSize)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.PageSuccess(c, messages, total, page, pageSize)
}

func (h *AdminChatHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的消息 ID")
		return
	}
	if err := h.svc.AdminDeleteMessage(uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}