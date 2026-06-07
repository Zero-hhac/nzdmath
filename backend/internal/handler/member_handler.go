package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	svc *service.MemberService
}

func NewMemberHandler(svc *service.MemberService) *MemberHandler {
	return &MemberHandler{
		svc: svc,
	}
}

func (h *MemberHandler) GetProfile(c *gin.Context) {
	userId := c.GetUint("user_id")

	profile, err := h.svc.GetProfile(userId)
	if err != nil {
		response.Fail(c, 400, "获取用户信息失败")
		return
	}
	response.Success(c, profile)
}

func (h *MemberHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Bio      string `json:"bio"`
		Email    string `json:"email"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}

	err = h.svc.UpdateProfile(userID, req.Nickname, req.Avatar, req.Bio, req.Email)
	if err != nil {
		response.Fail(c, 500, "更新用户信息失败")
		return
	}
	response.Success(c, nil)
}
