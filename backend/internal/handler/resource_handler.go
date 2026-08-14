package handler

import (
	"fmt"
	"math-top/internal/response"
	"math-top/internal/service"
	"math-top/internal/utils"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ResourceHandler struct {
	svc *service.ResourceService
}

func NewResourceHandler(svc *service.ResourceService) *ResourceHandler {
	return &ResourceHandler{svc: svc}
}

func (h *ResourceHandler) List(c *gin.Context) {
	resource, err := h.svc.ListResources()
	if err != nil {
		response.Fail(c, 500, "获取资源列表失败")
		return
	}
	response.Success(c, resource)
}

func (h *ResourceHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的文件 ID")
		return
	}
	resource, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Fail(c, 404, "无效的文件 ID")
		return
	}
	if resource.Status != 1 {
		response.Fail(c, 404, "资源不存在")
		return
	}
	h.svc.IncrementView(resource.ID)
	response.Success(c, resource)
}

func (h *ResourceHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "上传文件失败")
		return
	}
	title := c.PostForm("title")
	summary := c.PostForm("summary")
	category := c.PostForm("category")

	if title == "" {
		title = file.Filename
	}

	uploaderID := c.GetUint("user_id")
	if uploaderID == 0 {
		response.Fail(c, 401, "请先登录")
		return
	}

	resource, err := h.svc.UploadFile(
		file,
		"",
		summary,
		title,
		category,
		uploaderID,
	)
	if err != nil {
		response.Fail(c, 400, fmt.Sprintf("上传文件失败: %v", err))
		return
	}
	response.Success(c, gin.H{"id": resource.ID})
}

func (h *ResourceHandler) Download(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.Fail(c, 400, "无效的资源 ID")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效的资源 ID")
		return
	}

	resource, err := h.svc.GetForDownload(id)
	if err != nil {
		response.Fail(c, 404, "资源不存在")
		return
	}

	// 若为已登录用户下载，记录下载日志
	if userID := h.resolveUserID(c); userID > 0 {
		_ = h.svc.RecordDownload(userID, uint(id), c.ClientIP(), c.Request.UserAgent())
	}

	ext := strings.ToLower(filepath.Ext(resource.FileName))
	inlineExts := map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".txt": true}
	disposition := "attachment"
	if inlineExts[ext] {
		disposition = "inline"
	}

	safeName := strings.NewReplacer("\r", "", "\n", "", `"`, "").Replace(resource.FileName)
	c.Header("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disposition, safeName))
	c.File(resource.FilePath)
}

func (h *ResourceHandler) resolveUserID(c *gin.Context) uint {
	token := ""
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token = strings.TrimPrefix(auth, "Bearer ")
	} else if qToken := c.Query("token"); qToken != "" {
		token = qToken
	}
	if token == "" {
		return 0
	}
	claims, err := utils.ParseToken(token)
	if err != nil {
		return 0
	}
	return claims.UserID
}

func (h *ResourceHandler) MyDownloads(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	logs, total, err := h.svc.ListMyDownloads(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取下载历史失败")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	response.PageSuccess(c, logs, total, page, pageSize)
}
