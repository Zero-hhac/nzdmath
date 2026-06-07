package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminCommentHandler struct {
	svc *service.CommentService
}

func NewAdminCommentHandler(svc *service.CommentService) *AdminCommentHandler {
	return &AdminCommentHandler{svc: svc}
}

func (h *AdminCommentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	targetType := c.Query("target_type")
	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}
	comments, total, err := h.svc.ListWithFilter(page, pageSize, targetType, status)
	if err != nil {
		response.Fail(c, 500, "获取评论列表失败")
		return
	}
	comments = h.svc.FillUserInfo(comments)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	response.PageSuccess(c, comments, total, page, pageSize)
}

func (h *AdminCommentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的评论 ID")
		return
	}
	if err := h.svc.AdminDelete(uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

type SetCommentStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

func (h *AdminCommentHandler) SetStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的评论 ID")
		return
	}
	var req SetCommentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.SetStatus(uint(id), req.Status); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}