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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	news, total, err := h.svc.ListNews(page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取新闻资讯失败")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	response.PageSuccess(c, news, total, page, pageSize)
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
