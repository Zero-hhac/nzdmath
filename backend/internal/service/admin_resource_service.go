package service

import (
	"errors"
	"math-top/internal/config"
	"math-top/internal/dto"
	"math-top/internal/model"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

type AdminResourceService struct {
	db *gorm.DB
}

func NewAdminResourceService(db *gorm.DB) *AdminResourceService {
	return &AdminResourceService{db: db}
}

// archiveBase 返回归档目录（下线文件的移入目标）：{uploadDir}/archive/resources/<相对路径>。
func archiveBase() string {
	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}
	return filepath.Join(uploadBase, "archive", "resources")
}

// moveToArchive 将公开目录下的物理文件移入归档目录（#12：下线即不可下载）。
// 返回归档后的完整路径；失败时返回错误（不阻断数据库操作，但记录日志）。
func moveToArchive(srcPath string) (string, error) {
	if srcPath == "" {
		return "", nil
	}
	absSrc, err := filepath.Abs(srcPath)
	if err != nil {
		return "", err
	}
	baseAbs, err := filepath.Abs(archiveBase())
	if err != nil {
		return "", err
	}
	// 仅处理公开目录（uploads/resources）内的文件，防止把归档文件再次归档
	if !strings.Contains(filepath.ToSlash(absSrc), "/uploads/resources/") {
		return "", nil
	}
	rel := strings.TrimPrefix(filepath.ToSlash(absSrc), filepath.ToSlash(filepath.Dir(baseAbs)))
	dest := filepath.Join(baseAbs, filepath.FromSlash(strings.TrimPrefix(rel, "/")))
	if err := os.MkdirAll(filepath.Dir(dest), 0o775); err != nil {
		return "", err
	}
	if err := os.Rename(absSrc, dest); err != nil {
		// 目标已存在等场景：尝试复制后删除
		return "", err
	}
	return dest, nil
}

// restoreFromArchive 将归档文件移回公开目录（#12：恢复上架后可再次下载）。
func restoreFromArchive(archivedPath string, originalPath string) error {
	if archivedPath == "" || originalPath == "" {
		return nil
	}
	if _, err := os.Stat(archivedPath); err != nil {
		return nil // 归档文件不存在则跳过（如历史数据无文件）
	}
	if err := os.MkdirAll(filepath.Dir(originalPath), 0o775); err != nil {
		return err
	}
	return os.Rename(archivedPath, originalPath)
}

func (s *AdminResourceService) List(q dto.AdminListQuery) ([]model.Resource, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}

	tx := s.db.Model(&model.Resource{})
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.Keyword != "" {
		tx = tx.Where("title LIKE ?", "%"+q.Keyword+"%")
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var resources []model.Resource
	if err := tx.Order("id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&resources).Error; err != nil {
		return nil, 0, err
	}
	return resources, total, nil
}

func (s *AdminResourceService) Get(id uint) (*model.Resource, error) {
	var resource model.Resource
	if err := s.db.First(&resource, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("资源不存在")
		}
		return nil, err
	}
	return &resource, nil
}

// Update 更新资源；#12：状态 1→0 时把物理文件移出公开目录归档，0→1 时移回。
func (s *AdminResourceService) Update(id uint, req dto.AdminUpdateResourceRequest) error {
	var current model.Resource
	if err := s.db.First(&current, id).Error; err != nil {
		return errors.New("资源不存在")
	}

	res := s.db.Model(&model.Resource{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title":       req.Title,
		"summary":     req.Summary,
		"category":    req.Category,
		"cover_url":   req.CoverURL,
		"status":      req.Status,
		"is_featured": req.IsFeatured,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("资源不存在")
	}

	// 物理文件随状态迁移：1→0 下线归档，0→1 恢复移回
	if current.Status == 1 && req.Status == 0 {
		if dest, err := moveToArchive(current.FilePath); err == nil && dest != "" {
			s.db.Model(&model.Resource{}).Where("id = ?", id).Update("file_path", dest)
		}
	} else if current.Status == 0 && req.Status == 1 {
		// 恢复：把归档路径还原为公开路径
		uploadBase := config.GlobalConfig.Storage.UploadDir
		if uploadBase == "" {
			uploadBase = "./storage/uploads"
		}
		if strings.Contains(filepath.ToSlash(current.FilePath), "/archive/resources/") {
			rel := strings.TrimPrefix(filepath.ToSlash(current.FilePath), filepath.ToSlash(archiveBase()))
			original := filepath.Join(uploadBase, "resources", filepath.FromSlash(strings.TrimPrefix(rel, "/")))
			if err := restoreFromArchive(current.FilePath, original); err == nil {
				s.db.Model(&model.Resource{}).Where("id = ?", id).Update("file_path", original)
			}
		}
	}
	return nil
}

// Delete 删除资源；#12：同步删除物理文件。
func (s *AdminResourceService) Delete(id uint) error {
	var resource model.Resource
	if err := s.db.First(&resource, id).Error; err != nil {
		return errors.New("资源不存在")
	}
	res := s.db.Delete(&model.Resource{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("资源不存在")
	}
	// 删除物理文件（公开目录或归档目录均可清理）
	if resource.FilePath != "" {
		if err := os.Remove(resource.FilePath); err != nil && !os.IsNotExist(err) {
			return nil // 文件清理失败不阻断删除流程
		}
	}
	return nil
}
