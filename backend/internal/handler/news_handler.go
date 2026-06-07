package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
	svc *service.NewsService
}

func NewNewsHandler(svc *service.NewsService) *NewsHandler {
	return &NewsHandler{
		svc: svc,
	}
}

func (h *NewsHandler) List(c *gin.Context) {
	news, err := h.svc.ListNews()
	if err != nil {
		response.Fail(c, 500, "获取新闻资讯失败")
		return
	}
	response.Success(c, news)
}

func (h *NewsHandler) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的新闻 ID")
		return
	}

	news, err := h.svc.GetNewsByID(id)
	if err != nil {
		response.Fail(c, 404, "新闻资讯不存在")
		return
	}
	response.Success(c, news)
}
