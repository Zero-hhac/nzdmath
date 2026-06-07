package handler

import (
	"math-top/internal/dto"
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminNewsHandler struct {
	svc *service.AdminNewsService
}

func NewAdminNewsHandler(svc *service.AdminNewsService) *AdminNewsHandler {
	return &AdminNewsHandler{svc: svc}
}

func (h *AdminNewsHandler) List(c *gin.Context) {
	var q dto.AdminListQuery
	_ = c.ShouldBindQuery(&q)

	news, total, err := h.svc.List(q)
	if err != nil {
		response.Fail(c, 500, "获取资讯列表失败")
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
	response.PageSuccess(c, news, total, page, pageSize)
}

func (h *AdminNewsHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的资讯 ID")
		return
	}
	news, err := h.svc.Get(uint(id))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	response.Success(c, news)
}

func (h *AdminNewsHandler) Create(c *gin.Context) {
	adminID := c.GetUint("user_id")
	var req dto.AdminCreateNewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	news, err := h.svc.Create(adminID, req)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, news)
}

func (h *AdminNewsHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的资讯 ID")
		return
	}
	var req dto.AdminUpdateNewsRequest
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

func (h *AdminNewsHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的资讯 ID")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}
