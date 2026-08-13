package handler

import (
	"errors"
	"math-top/internal/response"
	"math-top/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FavoriteHandler struct {
	svc *service.FavoriteService
}

func NewFavoriteHandler(svc *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{svc: svc}
}

func (h *FavoriteHandler) AddFavorite(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		TargetID   uint   `json:"target_id"`
		TargetType string `json:"target_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	err := h.svc.AddFavorite(userID, req.TargetID, req.TargetType)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFavoriteInvalidType), errors.Is(err, service.ErrFavoriteDuplicate):
			response.Fail(c, 400, err.Error())
		default:
			response.Fail(c, 500, "添加收藏失败")
		}
		return
	}
	response.Success(c, nil)
}

func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID := c.GetUint("user_id")
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的收藏 ID")
		return
	}

	if err := h.svc.RemoveFavorite(uint(id), userID); err != nil {
		switch {
		case errors.Is(err, service.ErrFavoriteForbidden):
			response.Fail(c, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrFavoriteNotFound):
			response.Fail(c, http.StatusNotFound, err.Error())
		default:
			response.Fail(c, http.StatusInternalServerError, "删除收藏失败")
		}
		return
	}
	response.Success(c, nil)
}

func (h *FavoriteHandler) ListFavorites(c *gin.Context) {
	userID := c.GetUint("user_id")
	favorites, err := h.svc.ListFavorites(userID)
	if err != nil {
		response.Fail(c, 500, "获取收藏列表失败")
		return
	}
	response.Success(c, favorites)
}
