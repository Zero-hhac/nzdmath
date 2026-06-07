package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ShowcaseHandler struct {
	svc *service.ShowcaseService
}

func NewShowcaseHandler(svc *service.ShowcaseService) *ShowcaseHandler {
	return &ShowcaseHandler{svc: svc}
}

func (h *ShowcaseHandler) List(c *gin.Context) {
	field := c.Query("field")
	competition := c.Query("competition")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	showcases, total, err := h.svc.ListShowcases(field, competition, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取作品列表失败")
		return
	}
	response.PageSuccess(c, showcases, total, page, pageSize)
}

func (h *ShowcaseHandler) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的作品 ID")
		return
	}

	showcase, err := h.svc.GetShowcase(uint(id))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	response.Success(c, showcase)
}
