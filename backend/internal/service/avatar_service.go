package service

import (
	"errors"
	"fmt"
	"io"
	"math-top/internal/config"
	"math-top/internal/model"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AvatarService struct {
	db *gorm.DB
}

func NewAvatarService(db *gorm.DB) *AvatarService {
	return &AvatarService{db: db}
}

var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var allowedImageContentTypes = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
}

func (s *AvatarService) Upload(userID uint, file *multipart.FileHeader) (string, error) {
	if file.Size > 5*1024*1024 {
		return "", errors.New("头像文件不能超过 5MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		return "", errors.New("只支持 jpg/jpeg/png/gif/webp 格式")
	}

	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}
	now := time.Now()
	dir := filepath.Join(uploadBase, "avatars", now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(dir, 0775); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	filename := fmt.Sprintf("u%d_%d%s", userID, now.UnixNano(), ext)
	savePath := filepath.Join(dir, filename)

	src, err := file.Open()
	if err != nil {
		return "", errors.New("打开文件失败")
	}
	defer src.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", errors.New("读取文件失败")
	}
	if !allowedImageContentTypes[http.DetectContentType(head[:n])] {
		return "", errors.New("文件内容不是有效的图片")
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", errors.New("读取文件失败")
	}

	dst, err := os.Create(savePath)
	if err != nil {
		return "", errors.New("保存文件失败")
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return "", errors.New("写入文件失败")
	}

	url := fmt.Sprintf("/uploads/avatars/%s/%s/%s", now.Format("2006"), now.Format("01"), filename)
	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Update("avatar", url).Error; err != nil {
		return "", errors.New("更新头像失败")
	}
	return url, nil
}