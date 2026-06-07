package handler

import (
	"math-top/internal/dto"
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminShowcaseHandler struct {
	svc *service.AdminShowcaseService
}

func NewAdminShowcaseHandler(svc *service.AdminShowcaseService) *AdminShowcaseHandler {
	return &AdminShowcaseHandler{svc: svc}
}

func (h *AdminShowcaseHandler) List(c *gin.Context) {
	var q dto.AdminListQuery
	_ = c.ShouldBindQuery(&q)

	showcases, total, err := h.svc.List(q)
	if err != nil {
		response.Fail(c, 500, "获取作品列表失败")
		return
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	response.PageSuccess(c, showcases, total, page, pageSize)
}

func (h *AdminShowcaseHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的作品 ID")
		return
	}
	showcase, err := h.svc.Get(uint(id))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	response.Success(c, showcase)
}

func (h *AdminShowcaseHandler) Create(c *gin.Context) {
	var req dto.AdminCreateShowcaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	showcase, err := h.svc.Create(req)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, showcase)
}

func (h *AdminShowcaseHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的作品 ID")
		return
	}
	var req dto.AdminUpdateShowcaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(uint(id), req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminShowcaseHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的作品 ID")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}
