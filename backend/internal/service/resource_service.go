package service

import (
	"errors"
	"fmt"
	"math-top/internal/config"
	"math-top/internal/dto"
	"math-top/internal/model"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ResourceService struct {
	db *gorm.DB
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
	s.db.Model(&model.Resource{}).Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + 1"))
	return &resource, nil
}

func (s *ResourceService) UploadFile(fileName string, fileSize int64, fileType string, fileData []byte, coverURL string, summary string, title string, category string, uploaderID uint) error {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}
	uploadDir := filepath.Join(uploadBase, "resources", year, month)
	if err := os.MkdirAll(uploadDir, 0775); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	ext := filepath.Ext(fileName)
	uniqueName := fmt.Sprintf("%d_%s", now.UnixNano(), fileName)
	savePath := filepath.Join(uploadDir, uniqueName)
	if err := os.WriteFile(savePath, fileData, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	if uploaderID == 0 {
		uploaderID = 1
	}

	resource := model.Resource{
		Title:      title,
		Summary:    summary,
		Category:   category,
		FileName:   fileName,
		FilePath:   savePath,
		FileSize:   fileSize,
		FileType:   fileType,
		FileExt:    strings.ToLower(ext),
		CoverURL:   coverURL,
		Status:     1,
		CreatedBy:  uploaderID,
	}

	if err := s.db.Create(&resource).Error; err != nil {
		os.Remove(savePath)
		return fmt.Errorf("数据库记录创建失败: %v", err)
	}
	return nil
}

func (s *ResourceService) IncrementView(id uint) {
	s.db.Model(&model.Resource{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))
}

func (s *ResourceService) RecordDownload(userID uint, resourceID uint, ip, userAgent string) error {
	log := model.DownloadLog{
		UserID:       userID,
		ResourceID:   resourceID,
		IP:           ip,
		UserAgent:    userAgent,
	}
	return s.db.Create(&log).Error
}

func (s *ResourceService) ListMyDownloads(userID uint, page, pageSize int) ([]model.DownloadLog, int64, error) {
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
	return logs, total, nil
}

func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

var _ = parseUint
var _ = errors.New