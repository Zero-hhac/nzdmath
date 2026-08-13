package handler

import (
	"math-top/internal/dto"
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminEventHandler struct {
	svc *service.AdminEventService
}

func NewAdminEventHandler(svc *service.AdminEventService) *AdminEventHandler {
	return &AdminEventHandler{svc: svc}
}

func (h *AdminEventHandler) List(c *gin.Context) {
	var q dto.AdminListQuery
	_ = c.ShouldBindQuery(&q)

	events, total, err := h.svc.List(q)
	if err != nil {
		response.Fail(c, 500, "获取活动列表失败")
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
	response.PageSuccess(c, events, total, page, pageSize)
}

func (h *AdminEventHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	event, err := h.svc.Get(uint(id))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	response.Success(c, event)
}

func (h *AdminEventHandler) Create(c *gin.Context) {
	adminID := c.GetUint("user_id")
	var req dto.AdminCreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	event, err := h.svc.Create(adminID, req)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, event)
}

func (h *AdminEventHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	var req dto.AdminUpdateEventRequest
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

func (h *AdminEventHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminEventHandler) ToggleFeature(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	if err := h.svc.ToggleFeature(uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// SetExpired 管理员手动设置活动过期状态（公开列表下线/禁止报名，名单与记录仍可见）
func (h *AdminEventHandler) SetExpired(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的活动 ID")
		return
	}
	var req struct {
		IsExpired bool `json:"is_expired"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.SetExpired(uint(id), req.IsExpired); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"is_expired": req.IsExpired})
}
