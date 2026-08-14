package service

import (
	"errors"
	"math-top/internal/model"

	"gorm.io/gorm"
)

type ShowcaseService struct {
	db *gorm.DB
}

func NewShowcaseService(db *gorm.DB) *ShowcaseService {
	return &ShowcaseService{db: db}
}

func (s *ShowcaseService) ListShowcases(field, competition, keyword string, page, pageSize int) ([]model.Showcase, int64, error) {
	var showcases []model.Showcase
	var total int64

	query := s.db.Model(&model.Showcase{}).Where("status = ?", 1)

	if field != "" {
		query = query.Where("field = ?", field)
	}
	if competition != "" {
		query = query.Where("competition LIKE ?", "%"+competition+"%")
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	if err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&showcases).Error; err != nil {
		return nil, 0, err
	}

	return showcases, total, nil
}

func (s *ShowcaseService) GetShowcase(id uint) (*model.Showcase, error) {
	var showcase model.Showcase
	if err := s.db.Where("id = ? AND status = ?", id, 1).First(&showcase).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("作品不存在")
		}
		return nil, err
	}

	// 浏览量 +1
	s.db.Model(&showcase).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	return &showcase, nil
}
