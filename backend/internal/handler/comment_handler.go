package handler

import (
	"errors"
	"math-top/internal/response"
	"math-top/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

type CreateCommentRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Rating     int    `json:"rating"`
	ParentID   *uint  `json:"parent_id"`
}

func (h *CommentHandler) ListByTarget(c *gin.Context) {
	targetType := c.Query("target_type")
	targetIDStr := c.Query("target_id")
	if targetType == "" || targetIDStr == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	targetID, err := strconv.ParseUint(targetIDStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的目标 ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	comments, total, err := h.svc.ListByTarget(targetType, uint(targetID), page, pageSize)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	for i := range comments {
		comments[i].Replies = h.svc.FillUserInfo(comments[i].Replies)
	}
	comments = h.svc.FillUserInfoAsSlice(comments)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	response.PageSuccess(c, comments, total, page, pageSize)
}

// ListReplies 分页加载某条父评论的回复（#24）
func (h *CommentHandler) ListReplies(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的评论 ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	replies, total, hasMore, err := h.svc.ListReplies(uint(id), page, pageSize)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	response.PageSuccess(c, gin.H{
		"replies":  replies,
		"has_more": hasMore,
	}, total, page, pageSize)
}

func (h *CommentHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	comment, err := h.svc.Create(userID, service.CreateCommentParams{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Content:    req.Content,
		Rating:     req.Rating,
		ParentID:   req.ParentID,
	})
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, comment)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的评论 ID")
		return
	}
	if err := h.svc.Delete(userID, uint(id)); err != nil {
		switch {
		case errors.Is(err, service.ErrCommentForbidden):
			response.Fail(c, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrCommentNotFound):
			response.Fail(c, http.StatusNotFound, err.Error())
		default:
			response.Fail(c, http.StatusBadRequest, err.Error())
		}
		return
	}
	response.Success(c, nil)
}

func (h *CommentHandler) ToggleLike(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的评论 ID")
		return
	}
	liked, err := h.svc.ToggleLike(userID, uint(id))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"liked": liked})
}
