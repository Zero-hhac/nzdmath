package handler

import (
	"fmt"
	"io"
	"math-top/internal/response"
	"math-top/internal/service"
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
	openedFile, err := file.Open()
	if err != nil {
		response.Fail(c, 500, "打开文件失败")
		return
	}
	defer openedFile.Close()

	fileData, err := io.ReadAll(openedFile)
	if err != nil {
		response.Fail(c, 500, "读取文件失败")
		return
	}

	uploaderID := c.GetUint("user_id")
	if uploaderID == 0 {
		uploaderID = 1
	}

	err = h.svc.UploadFile(
		file.Filename,
		file.Size,
		file.Header.Get("Content-Type"),
		fileData,
		"",
		summary,
		title,
		category,
		uploaderID,
	)
	if err != nil {
		response.Fail(c, 500, fmt.Sprintf("上传文件失败: %v", err))
		return
	}
	response.Success(c, nil)
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

	ext := strings.ToLower(filepath.Ext(resource.FileName))
	inlineExts := map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".txt": true}
	disposition := "attachment"
	if inlineExts[ext] {
		disposition = "inline"
	}

	c.Header("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, disposition, resource.FileName))
	c.File(resource.FilePath)
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