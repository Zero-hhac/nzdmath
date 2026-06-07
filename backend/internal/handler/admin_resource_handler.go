package handler

import (
	"math-top/internal/dto"
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminResourceHandler struct {
	svc *service.AdminResourceService
}

func NewAdminResourceHandler(svc *service.AdminResourceService) *AdminResourceHandler {
	return &AdminResourceHandler{svc: svc}
}

func (h *AdminResourceHandler) List(c *gin.Context) {
	var q dto.AdminListQuery
	_ = c.ShouldBindQuery(&q)

	resources, total, err := h.svc.List(q)
	if err != nil {
		response.Fail(c, 500, "获取资源列表失败")
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
	response.PageSuccess(c, resources, total, page, pageSize)
}

func (h *AdminResourceHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的资源 ID")
		return
	}
	resource, err := h.svc.Get(uint(id))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	response.Success(c, resource)
}

func (h *AdminResourceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的资源 ID")
		return
	}
	var req dto.AdminUpdateResourceRequest
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

func (h *AdminResourceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的资源 ID")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}
