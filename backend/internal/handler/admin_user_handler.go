package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct {
	svc *service.AdminUserService
}

func NewAdminUserHandler(svc *service.AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{svc: svc}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	department := c.Query("department")
	incomplete := c.Query("incomplete") == "1"
	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}
	users, total, err := h.svc.List(page, pageSize, keyword, status, department, incomplete)
	if err != nil {
		response.Fail(c, 500, "获取用户列表失败")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	response.PageSuccess(c, users, total, page, pageSize)
}

type SetStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

func (h *AdminUserHandler) SetStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的用户 ID")
		return
	}
	var req SetStatusRequest
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

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的用户 ID")
		return
	}
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := h.svc.ResetPassword(uint(id), req.NewPassword); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的用户 ID")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

type BatchSetStatusRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status int    `json:"status" binding:"oneof=0 1"`
}

func (h *AdminUserHandler) BatchSetStatus(c *gin.Context) {
	var req BatchSetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	affected, err := h.svc.BatchSetStatus(req.IDs, req.Status)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"affected": affected})
}

type BatchResetPasswordRequest struct {
	IDs         []uint `json:"ids" binding:"required,min=1"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AdminUserHandler) BatchResetPassword(c *gin.Context) {
	var req BatchResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	affected, err := h.svc.BatchResetPassword(req.IDs, req.NewPassword)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"affected": affected})
}

type BatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

func (h *AdminUserHandler) BatchDelete(c *gin.Context) {
	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	affected, err := h.svc.BatchDelete(req.IDs)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"affected": affected})
}

func (h *AdminUserHandler) Export(c *gin.Context) {
	department := c.Query("department")
	f, err := h.svc.ExportExcel(department)
	if err != nil {
		response.Fail(c, 500, "导出失败")
		return
	}

	name := "会员名单"
	if department != "" {
		name = "会员名单-" + department
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="members.xlsx"; filename*=UTF-8''`+url.PathEscape(name+".xlsx"))
	_ = f.Write(c.Writer)
}