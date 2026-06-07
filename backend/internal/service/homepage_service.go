package service

import (
	"context"
	"math-top/internal/cache"
	"math-top/internal/consts"
	"math-top/internal/model"
	"sync"
	"time"

	"gorm.io/gorm"
)

type HomepageService struct {
	db    *gorm.DB
	cache *cache.Cache
}

type HomepageData struct {
	RecentEvents    []model.Event    `json:"recent_events"`
	FeaturedEvents  []model.Event    `json:"featured_events"`
	RecentNews      []model.News     `json:"recent_news"`
	FeaturedNews    []model.News     `json:"featured_news"`
	FeaturedResources []model.Resource `json:"featured_resources"`
	RecentShowcases []model.Showcase  `json:"recent_showcases"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func NewHomepageService(db *gorm.DB, c *cache.Cache) *HomepageService {
	return &HomepageService{db: db, cache: c}
}

func (s *HomepageService) Get() (*HomepageData, error) {
	ctx := context.Background()
	version := s.cache.GetInt(ctx, "cache:homepage:version")
	cacheKey := "cache:homepage:v" + intToStr(version)

	var cached HomepageData
	if s.cache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	data := &HomepageData{UpdatedAt: time.Now()}

	var wg sync.WaitGroup
	wg.Add(6)
	go func() { defer wg.Done(); s.db.Select("id, title, summary, category, location, start_time, end_time, cover_url, is_featured, status, created_at").Where("status = ?", 1).Order("id desc").Limit(8).Find(&data.RecentEvents) }()
	go func() { defer wg.Done(); s.db.Select("id, title, summary, category, location, start_time, end_time, cover_url, is_featured, status, created_at").Where("status = ? AND is_featured = ?", 1, true).Order("id desc").Limit(4).Find(&data.FeaturedEvents) }()
	go func() { defer wg.Done(); s.db.Select("id, title, summary, category, tag, cover_url, is_featured, status, published_at, created_at").Where("status = ?", 1).Order("published_at desc, id desc").Limit(8).Find(&data.RecentNews) }()
	go func() { defer wg.Done(); s.db.Select("id, title, summary, category, tag, cover_url, is_featured, status, published_at, created_at").Where("status = ? AND is_featured = ?", 1, true).Order("id desc").Limit(4).Find(&data.FeaturedNews) }()
	go func() { defer wg.Done(); s.db.Select("id, title, summary, category, cover_url, view_count, download_count, like_count, is_featured, file_ext, file_size, created_at").Where("status = ? AND is_featured = ?", 1, true).Order("id desc").Limit(4).Find(&data.FeaturedResources) }()
	go func() { defer wg.Done(); s.db.Select("id, title, author, field, competition, summary, cover_url, view_count, status, created_at").Where("status = ?", 1).Order("view_count desc, id desc").Limit(4).Find(&data.RecentShowcases) }()
	wg.Wait()

	s.cache.Set(ctx, cacheKey, data, time.Duration(consts.CacheTTLHomepage)*time.Second)
	return data, nil
}

func (s *HomepageService) Invalidate() {
	s.cache.Incr(context.Background(), "cache:homepage:version")
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}