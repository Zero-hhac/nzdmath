package handler

import (
	"errors"
	"fmt"
	"math-top/internal/middleware"
	"math-top/internal/model"
	"math-top/internal/response"
	"math-top/internal/utils"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DirectFileHandler 私聊附件下载接口：/uploads/direct/*filepath
// 不再走静态挂载，改为按私信记录做归属校验：
//   - 仅会话会员（发送方/接收方）与管理员可下载；
//   - 第三方即使持有完整地址也返回 403；
//   - 同时兼容用户 token 与管理员 token（管理端面板用 admin token 拉取）。
type DirectFileHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewDirectFileHandler(db *gorm.DB, rdb *redis.Client) *DirectFileHandler {
	return &DirectFileHandler{db: db, rdb: rdb}
}

// DownloadFile 校验归属后返回文件内容。
func (h *DirectFileHandler) DownloadFile(c *gin.Context) {
	filePath := c.Param("filepath")
	if filePath == "" {
		response.Fail(c, http.StatusNotFound, "文件不存在")
		return
	}
	fileURL := "/uploads/direct" + filePath

	claims, isAdmin, err := h.authenticate(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "未登录或登录已过期")
		return
	}

	var msg model.DirectMessage
	if err := h.db.Where("file_url = ?", fileURL).Order("id desc").First(&msg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "文件不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "查询文件失败")
		return
	}

	// 归属检查：管理员放行；会员仅限会话本人
	if !isAdmin && claims.UserID != msg.UserID {
		response.Fail(c, http.StatusForbidden, "无权访问该文件")
		return
	}

	ext := strings.ToLower(filepath.Ext(msg.FileName))
	inlineExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".pdf": true}
	disposition := "attachment"
	if inlineExts[ext] {
		disposition = "inline"
	}
	safeName := strings.NewReplacer("\r", "", "\n", "", "\"", "").Replace(msg.FileName)
	c.Header("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"", disposition, safeName))
	c.File(msg.FilePath)
}

// authenticate 解析 Authorization: Bearer <token>，优先按用户 token 校验，
// 失败再按管理员 token 校验；返回 claims 与是否管理员身份。
func (h *DirectFileHandler) authenticate(c *gin.Context) (*utils.MyClaims, bool, error) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return nil, false, errors.New("缺少 Authorization 头")
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, false, errors.New("Authorization 格式有误")
	}
	token := parts[1]

	if claims, err := middleware.AuthenticateToken(h.rdb, token, middleware.UserTokenPrefix); err == nil {
		isAdmin := claims.Role == "admin" || claims.Role == "super_admin"
		return claims, isAdmin, nil
	}
	if claims, err := middleware.AuthenticateToken(h.rdb, token, middleware.AdminTokenPrefix); err == nil {
		isAdmin := claims.Role == "admin" || claims.Role == "super_admin"
		return claims, isAdmin, nil
	}
	return nil, false, errors.New("token 无效或者已过期")
}
