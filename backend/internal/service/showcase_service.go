package service

import (
	"context"
	"errors"
	"fmt"
	"math-top/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ShowcaseService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewShowcaseService(db *gorm.DB, rdb *redis.Client) *ShowcaseService {
	return &ShowcaseService{db: db, rdb: rdb}
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
	// #10：与其他列表一致，page_size 封顶 50
	if pageSize > 50 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	if err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&showcases).Error; err != nil {
		return nil, 0, err
	}

	return showcases, total, nil
}

// GetShowcase 作品详情；#20：浏览量按“IP+作品”维度在 Redis 24 小时去重后再自增，
// 防止同一用户刷新刷榜（首页按浏览量排序）。
func (s *ShowcaseService) GetShowcase(id uint, clientIP string) (*model.Showcase, error) {
	var showcase model.Showcase
	if err := s.db.Where("id = ? AND status = ?", id, 1).First(&showcase).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("作品不存在")
		}
		return nil, err
	}

	if s.rdb != nil && clientIP != "" {
		key := fmt.Sprintf("showcase_view:%d:%s", id, clientIP)
		// SETNX 成功表示该 IP 24 小时内首次访问，计数才 +1
		ok, err := s.rdb.SetNX(context.Background(), key, "1", 24*time.Hour).Result()
		if err == nil && ok {
			s.db.Model(&showcase).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
		}
	}

	return &showcase, nil
}
