package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"

	"github.com/gin-gonic/gin"
)

type AvatarHandler struct {
	svc *service.AvatarService
}

func NewAvatarHandler(svc *service.AvatarService) *AvatarHandler {
	return &AvatarHandler{svc: svc}
}

func (h *AvatarHandler) Upload(c *gin.Context) {
	userID := c.GetUint("user_id")
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "请上传文件")
		return
	}
	url, err := h.svc.Upload(userID, file)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"avatar": url})
}