package service

import (
	"errors"
	"fmt"
	"io"
	"math-top/internal/config"
	"math-top/internal/dto"
	"math-top/internal/model"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ResourceService struct {
	db *gorm.DB
}

const ResourceMaxFileSize = 50 * 1024 * 1024

var allowedResourceExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".zip": true, ".txt": true, ".md": true,
	".json": true, ".csv": true,
}

func NewResourceService(db *gorm.DB) *ResourceService {
	return &ResourceService{
		db: db,
	}
}

func (s *ResourceService) ListResources() ([]dto.ResourceListItem, error) {
	var resources []model.Resource
	err := s.db.Select("id, title, summary, category, cover_url, view_count, download_count, is_featured, file_ext, file_size, created_at").
		Where("status = ?", 1).
		Order("id desc").
		Find(&resources).Error
	if err != nil {
		return nil, err
	}

	result := make([]dto.ResourceListItem, 0, len(resources))
	for _, resource := range resources {
		result = append(result, dto.ResourceListItem{
			ID:            resource.ID,
			Title:         resource.Title,
			Summary:       resource.Summary,
			Category:      resource.Category,
			CoverURL:      resource.CoverURL,
			DownloadCount: resource.DownloadCount,
			ViewCount:     resource.ViewCount,
		})
	}
	return result, nil
}

func (s *ResourceService) GetResourceByID(id string) (*dto.ResourceDetail, error) {
	var resource model.Resource
	err := s.db.Where("id = ? AND status = ?", id, 1).First(&resource).Error
	if err != nil {
		return nil, err
	}

	return &dto.ResourceDetail{
		ID:            resource.ID,
		Title:         resource.Title,
		Summary:       resource.Summary,
		Category:      resource.Category,
		FileName:      resource.FileName,
		FileSize:      resource.FileSize,
		FileType:      resource.FileType,
		CoverURL:      resource.CoverURL,
		DownloadCount: resource.DownloadCount,
		ViewCount:     resource.ViewCount,
	}, nil
}

func (s *ResourceService) GetByID(id uint) (*model.Resource, error) {
	var resource model.Resource
	if err := s.db.First(&resource, id).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

func (s *ResourceService) GetForDownload(id uint64) (*model.Resource, error) {
	var resource model.Resource
	if err := s.db.Where("id = ? AND status = ?", id, 1).First(&resource).Error; err != nil {
		return nil, err
	}
	if strings.HasPrefix(resource.FilePath, "/uploads/") {
		uploadBase := config.GlobalConfig.Storage.UploadDir
		if uploadBase == "" {
			uploadBase = "./storage/uploads"
		}
		baseAbs, err := filepath.Abs(uploadBase)
		if err != nil {
			return nil, errors.New("上传目录配置无效")
		}
		full := filepath.Clean(filepath.Join(baseAbs, strings.TrimPrefix(resource.FilePath, "/uploads/")))
		if !strings.HasPrefix(full, baseAbs+string(os.PathSeparator)) {
			return nil, errors.New("非法文件路径")
		}
		resource.FilePath = full
	}
	s.db.Model(&model.Resource{}).Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + 1"))
	return &resource, nil
}

func (s *ResourceService) UploadFile(file *multipart.FileHeader, coverURL string, summary string, title string, category string, uploaderID uint) (*model.Resource, error) {
	if file == nil {
		return nil, errors.New("上传文件不能为空")
	}
	ext, err := validateResourceUpload(file.Filename, file.Size)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}
	uploadDir := filepath.Join(uploadBase, "resources", year, month)
	if err := os.MkdirAll(uploadDir, 0775); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}

	saveName := fmt.Sprintf("%d_%d%s", now.UnixNano(), uploaderID, ext)
	savePath := filepath.Join(uploadDir, saveName)

	src, err := file.Open()
	if err != nil {
		return nil, errors.New("打开文件失败")
	}
	defer src.Close()

	if err := validateUploadContent(ext, src); err != nil {
		return nil, err
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, errors.New("读取文件失败")
	}

	dst, err := os.Create(savePath)
	if err != nil {
		return nil, errors.New("保存文件失败")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(savePath)
		return nil, errors.New("写入文件失败")
	}

	resource := model.Resource{
		Title:     title,
		Summary:   summary,
		Category:  category,
		FileName:  filepath.Base(file.Filename),
		FilePath:  savePath,
		FileSize:  file.Size,
		FileType:  file.Header.Get("Content-Type"),
		FileExt:   strings.ToLower(ext),
		CoverURL:  coverURL,
		Status:    1,
		CreatedBy: uploaderID,
	}

	if err := s.db.Create(&resource).Error; err != nil {
		os.Remove(savePath)
		return nil, fmt.Errorf("数据库记录创建失败: %v", err)
	}
	return &resource, nil
}

func validateResourceUpload(fileName string, size int64) (string, error) {
	if size > ResourceMaxFileSize {
		return "", fmt.Errorf("文件不能超过 %dMB", ResourceMaxFileSize/(1024*1024))
	}
	ext := strings.ToLower(filepath.Ext(filepath.Base(fileName)))
	if !allowedResourceExts[ext] {
		return "", errors.New("不支持的文件格式")
	}
	return ext, nil
}

func (s *ResourceService) IncrementView(id uint) {
	s.db.Model(&model.Resource{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))
}

func (s *ResourceService) RecordDownload(userID uint, resourceID uint, ip, userAgent string) error {
	log := model.DownloadLog{
		UserID:     userID,
		ResourceID: resourceID,
		IP:         ip,
		UserAgent:  userAgent,
	}
	return s.db.Create(&log).Error
}

func (s *ResourceService) ListMyDownloads(userID uint, page, pageSize int) ([]dto.MyDownloadItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	var total int64
	if err := s.db.Model(&model.DownloadLog{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []model.DownloadLog
	if err := s.db.Where("user_id = ?", userID).
		Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	if len(logs) == 0 {
		return []dto.MyDownloadItem{}, total, nil
	}

	resIDs := make([]uint, 0, len(logs))
	for _, l := range logs {
		resIDs = append(resIDs, l.ResourceID)
	}
	var resources []model.Resource
	s.db.Select("id, title, file_name, file_size, cover_url").Where("id IN ?", resIDs).Find(&resources)
	rm := make(map[uint]model.Resource, len(resources))
	for _, r := range resources {
		rm[r.ID] = r
	}

	result := make([]dto.MyDownloadItem, 0, len(logs))
	for _, l := range logs {
		item := dto.MyDownloadItem{
			ID:         l.ID,
			ResourceID: l.ResourceID,
			IP:         l.IP,
			CreatedAt:  l.CreatedAt,
		}
		if r, ok := rm[l.ResourceID]; ok {
			item.ResourceTitle = r.Title
			item.FileName = r.FileName
			item.FileSize = r.FileSize
			item.CoverURL = r.CoverURL
		} else {
			item.ResourceTitle = fmt.Sprintf("资源 #%d（已下线）", l.ResourceID)
		}
		result = append(result, item)
	}
	return result, total, nil
}
