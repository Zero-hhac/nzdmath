package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"

	"github.com/gin-gonic/gin"
)

type HomepageHandler struct {
	svc *service.HomepageService
}

func NewHomepageHandler(svc *service.HomepageService) *HomepageHandler {
	return &HomepageHandler{svc: svc}
}

func (h *HomepageHandler) Get(c *gin.Context) {
	data, err := h.svc.Get()
	if err != nil {
		response.Fail(c, 500, "获取首页数据失败")
		return
	}
	response.Success(c, data)
}