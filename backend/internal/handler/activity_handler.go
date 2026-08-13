package handler

import (
	"math-top/internal/response"
	"math-top/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	svc *service.ActivityService
}

func NewActivityHandler(svc *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func (h *ActivityHandler) Track(c *gin.Context) {
	tokenString := ""
	authHeader := c.GetHeader("Authorization")
	if parts := strings.Split(authHeader, " "); len(parts) == 2 && parts[0] == "Bearer" {
		tokenString = parts[1]
	}
	h.svc.Track(c.Request.Context(), c.ClientIP(), tokenString)
	response.Success(c, nil)
}
