package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math-top/internal/cache"
	"math-top/internal/config"
	"math-top/internal/model"
	"math-top/internal/ws"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	ChatMessageTypeText   = "text"
	ChatMessageTypeImage  = "image"
	ChatMessageTypeFile   = "file"
	ChatMessageTypeSystem = "system"

	ChatMaxFileSize  = 5 * 1024 * 1024
	chatOnlineTTL    = 30 * time.Second
	chatRecallWindow = 2 * time.Minute

	cacheKeyStream      = "chat:stream"
	cacheKeyMaxID       = "chat:max_id"
	cacheKeyPresence    = "chat:presence"
	cacheKeyDeleted     = "chat:deleted"
	chatStreamCap       = 2000
	chatDeleteRetention = 10 * time.Minute
)

type ChatService struct {
	db        *gorm.DB
	cache     *cache.Cache
	broadcast func(event string, data interface{})
}

// SetBroadcast 注入广播回调（由 router 组装时挂上 WebSocket Hub）。
// event: ws.TypeMessage / ws.TypePresence / ws.TypeDelete；nil 表示不广播（如单测）。
func (s *ChatService) SetBroadcast(fn func(event string, data interface{})) {
	s.broadcast = fn
}

func (s *ChatService) emit(event string, data interface{}) {
	if s.broadcast != nil {
		s.broadcast(event, data)
	}
}

type ChatMessagesResult struct {
	Messages     []model.ChatMessage `json:"messages"`
	OnlineCount  int64               `json:"online_count"`
	DeletedIDs   []uint              `json:"deleted_ids"`
	DeletedAtMs  int64               `json:"deleted_at_ms"`
	HasMore      bool                `json:"has_more"`
	NextBeforeID uint                `json:"next_before_id,omitempty"`
}

func NewChatService(db *gorm.DB, c *cache.Cache) *ChatService {
	return &ChatService{db: db, cache: c}
}

func (s *ChatService) Join(userID uint) (int64, error) {
	now := time.Now()
	cutoff := now.Add(-chatOnlineTTL)

	user, err := s.getUser(userID)
	if err != nil {
		return 0, err
	}

	shouldAnnounce := true
	if _, found, ok := s.cache.ZScore(context.Background(), cacheKeyPresence, strconv.FormatUint(uint64(userID), 10)); ok {
		shouldAnnounce = !found
	}
	var presence model.ChatPresence
	err = s.db.Where("user_id = ?", userID).First(&presence).Error
	if err == nil {
		shouldAnnounce = !presence.IsOnline || presence.LastSeenAt.Before(cutoff)
		if err := s.db.Model(&model.ChatPresence{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
			"is_online":    true,
			"joined_at":    now,
			"last_seen_at": now,
		}).Error; err != nil {
			return 0, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		presence = model.ChatPresence{
			UserID:     userID,
			IsOnline:   true,
			JoinedAt:   now,
			LastSeenAt: now,
		}
		if err := s.db.Create(&presence).Error; err != nil {
			return 0, err
		}
	} else {
		return 0, err
	}

	if shouldAnnounce {
		announceKey := fmt.Sprintf("chat:join:announce:%d", userID)
		if s.cache.SetNX(context.Background(), announceKey, "1", 5*time.Minute) {
			msg := model.ChatMessage{
				UserID:      userID,
				MessageType: ChatMessageTypeSystem,
				Content:     fmt.Sprintf("%s 加入了聊天室", displayName(user)),
			}
			if err := s.db.Create(&msg).Error; err != nil {
				return 0, err
			}
			s.cacheMessage(msg)
			s.emit(ws.TypeMessage, s.fillUserInfo([]model.ChatMessage{msg})[0])
		}
	}
	s.markOnline(userID, now)

	return s.OnlineCount()
}

func (s *ChatService) Leave(userID uint) error {
	if s.cache != nil {
		s.cache.ZRem(context.Background(), cacheKeyPresence, strconv.FormatUint(uint64(userID), 10))
	}
	return s.db.Model(&model.ChatPresence{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"is_online":    false,
		"last_seen_at": time.Now(),
	}).Error
}

func (s *ChatService) ListMessages(userID uint, afterID, beforeID uint, limit int, afterDeleteMs int64) (*ChatMessagesResult, error) {
	if err := s.touchPresence(userID); err != nil {
		return nil, err
	}

	if afterID > 0 {
		// Incremental poll: try Redis first.
		messages, hit := s.readIncrementalFromCache(afterID, limit)
		if !hit {
			// Cache miss (afterID older than stream tail): fall back to DB.
			if err := s.db.Where("id > ?", afterID).Order("created_at asc, id asc").
				Limit(safeLimit(limit)).Find(&messages).Error; err != nil {
				return nil, err
			}
		}
		return &ChatMessagesResult{
			Messages:    s.fillUserInfo(messages),
			OnlineCount: s.onlineCountFromCache(),
			DeletedIDs:  s.deletedIDsFromCache(afterDeleteMs),
			DeletedAtMs: nowMs(),
		}, nil
	}

	// History uses an ID cursor and only returns a bounded window.
	var messages []model.ChatMessage
	query := s.db.Order("id desc")
	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}
	pageLimit := safeLimit(limit)
	if err := query.Limit(pageLimit + 1).Find(&messages).Error; err != nil {
		return nil, err
	}
	hasMore := len(messages) > pageLimit
	if hasMore {
		messages = messages[:pageLimit]
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
	nextBeforeID := uint(0)
	if len(messages) > 0 {
		nextBeforeID = messages[0].ID
	}

	return &ChatMessagesResult{
		Messages:     s.fillUserInfo(messages),
		OnlineCount:  s.onlineCountFromCache(),
		DeletedIDs:   s.deletedIDsFromCache(afterDeleteMs),
		DeletedAtMs:  nowMs(),
		HasMore:      hasMore,
		NextBeforeID: nextBeforeID,
	}, nil
}

func safeLimit(limit int) int {
	if limit < 1 || limit > 200 {
		return 100
	}
	return limit
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}

// readIncrementalFromCache returns messages with id > afterID from the Redis stream.
// Returns hit=false when the cache cannot answer (missing/evicted/backfill needed).
func (s *ChatService) readIncrementalFromCache(afterID uint, limit int) ([]model.ChatMessage, bool) {
	if s.cache == nil {
		return nil, false
	}
	ctx := context.Background()
	oldest, ok := s.cache.ZRangeByRank(ctx, cacheKeyStream, 0, 0)
	if !ok || len(oldest) == 0 {
		return nil, false
	}
	var oldestMessage model.ChatMessage
	if err := json.Unmarshal([]byte(oldest[0]), &oldestMessage); err != nil || (afterID+1 < oldestMessage.ID) {
		return nil, false
	}
	maxID := s.cache.GetInt(ctx, cacheKeyMaxID)
	if maxID <= 0 {
		return nil, false
	}
	members := s.cache.ZRangeByScore(ctx, cacheKeyStream, float64(afterID)+1, float64(maxID))

	if limit <= 0 || limit > 200 {
		limit = 100
	}
	if len(members) > limit {
		members = members[:limit]
	}

	messages := make([]model.ChatMessage, 0, len(members))
	for _, m := range members {
		var msg model.ChatMessage
		if err := json.Unmarshal([]byte(m), &msg); err == nil {
			messages = append(messages, msg)
		}
	}
	return messages, true
}

// deletedIDsFromCache returns message ids deleted after afterDeleteMs from Redis,
// then trims entries older than the retention window.
func (s *ChatService) deletedIDsFromCache(afterDeleteMs int64) []uint {
	if s.cache == nil {
		return nil
	}
	ctx := context.Background()
	min := float64(0)
	if afterDeleteMs > 0 {
		min = float64(afterDeleteMs) + 1
	}
	members := s.cache.ZRangeByScore(ctx, cacheKeyDeleted, min, float64(nowMs()))
	if len(members) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(members))
	for _, m := range members {
		if id, err := strconv.ParseUint(m, 10, 64); err == nil {
			ids = append(ids, uint(id))
		}
	}
	// Trim entries older than the retention window to keep the set bounded.
	cutoff := float64(time.Now().Add(-chatDeleteRetention).UnixMilli())
	s.cache.ZRemRangeByScore(ctx, cacheKeyDeleted, 0, cutoff)
	return ids
}

func (s *ChatService) DeleteMessage(userID uint, messageID uint) error {
	var msg model.ChatMessage
	if err := s.db.First(&msg, messageID).Error; err != nil {
		return errors.New("消息不存在或已被删除")
	}

	var user model.User
	s.db.First(&user, userID)
	isAdmin := user.Role == "admin" || user.Role == "super_admin" || user.Username == "admin"

	if !isAdmin {
		if msg.UserID != userID {
			return errors.New("只能撤回自己的消息")
		}
		if time.Since(msg.CreatedAt) > chatRecallWindow {
			return errors.New("超过 2 分钟，消息无法撤回")
		}
	}

	if err := s.db.Delete(&model.ChatMessage{}, messageID).Error; err != nil {
		return errors.New("撤回失败")
	}
	s.cacheDeletion(messageID)
	s.emit(ws.TypeDelete, []uint{messageID})
	return nil
}

func (s *ChatService) AdminDeleteMessage(messageID uint) error {
	res := s.db.Delete(&model.ChatMessage{}, messageID)
	if res.Error != nil {
		return errors.New("删除失败")
	}
	if res.RowsAffected == 0 {
		return errors.New("消息不存在或已被删除")
	}
	s.cacheDeletion(messageID)
	s.emit(ws.TypeDelete, []uint{messageID})
	return nil
}

func (s *ChatService) AdminListMessages(page, pageSize int) ([]model.ChatMessage, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	var total int64
	if err := s.db.Model(&model.ChatMessage{}).Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取消息总数失败")
	}
	var messages []model.ChatMessage
	if err := s.db.Order("created_at desc, id desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&messages).Error; err != nil {
		return nil, 0, errors.New("获取消息列表失败")
	}
	return s.fillUserInfo(messages), total, nil
}

func (s *ChatService) SendText(userID uint, content string) (*model.ChatMessage, error) {
	if _, err := s.getUser(userID); err != nil {
		return nil, err
	}
	if err := s.touchPresence(userID); err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("消息内容不能为空")
	}
	if len([]rune(content)) > 2000 {
		return nil, errors.New("消息内容不能超过 2000 字")
	}

	msg := &model.ChatMessage{
		UserID:      userID,
		MessageType: ChatMessageTypeText,
		Content:     content,
	}
	if err := s.db.Create(msg).Error; err != nil {
		return nil, errors.New("发送消息失败")
	}
	s.cacheMessage(*msg)

	filled := s.fillUserInfo([]model.ChatMessage{*msg})
	s.emit(ws.TypeMessage, filled[0])
	return &filled[0], nil
}

func (s *ChatService) SendFile(userID uint, file *multipart.FileHeader) (*model.ChatMessage, error) {
	if _, err := s.getUser(userID); err != nil {
		return nil, err
	}
	if err := s.touchPresence(userID); err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.New("请上传文件")
	}
	if file.Size > ChatMaxFileSize {
		return nil, errors.New("文件不能超过 5MB")
	}

	originalName := filepath.Base(file.Filename)
	ext := strings.ToLower(filepath.Ext(originalName))
	mime := file.Header.Get("Content-Type")
	if chatVideoExts[ext] || strings.HasPrefix(strings.ToLower(mime), "video/") {
		return nil, errors.New("不支持上传视频文件")
	}

	messageType := ChatMessageTypeFile
	if chatImageExts[ext] {
		messageType = ChatMessageTypeImage
	} else if !chatFileExts[ext] {
		return nil, errors.New("不支持的文件格式")
	}

	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}

	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	dir := filepath.Join(uploadBase, "chat", year, month)
	if err := os.MkdirAll(dir, 0775); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}

	saveName := fmt.Sprintf("u%d_%d%s", userID, now.UnixNano(), ext)
	savePath := filepath.Join(dir, saveName)

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

	if _, err := dst.ReadFrom(src); err != nil {
		os.Remove(savePath)
		return nil, errors.New("写入文件失败")
	}

	msg := &model.ChatMessage{
		UserID:      userID,
		MessageType: messageType,
		FileName:    originalName,
		FileURL:     fmt.Sprintf("/uploads/chat/%s/%s/%s", year, month, saveName),
		FilePath:    savePath,
		FileSize:    file.Size,
		FileExt:     ext,
		FileMime:    mime,
	}
	if err := s.db.Create(msg).Error; err != nil {
		os.Remove(savePath)
		return nil, errors.New("保存聊天消息失败")
	}
	s.cacheMessage(*msg)

	filled := s.fillUserInfo([]model.ChatMessage{*msg})
	s.emit(ws.TypeMessage, filled[0])
	return &filled[0], nil
}

func (s *ChatService) OnlineCount() (int64, error) {
	if s.cache != nil {
		ctx := context.Background()
		cutoff := float64(time.Now().Add(-chatOnlineTTL).UnixMilli())
		s.cache.ZRemRangeByScore(ctx, cacheKeyPresence, 0, cutoff)
		if count, ok := s.cache.ZCount(ctx, cacheKeyPresence, cutoff+1, float64(nowMs())); ok {
			return count, nil
		}
	}
	return s.computeOnlineCount()
}

func (s *ChatService) computeOnlineCount() (int64, error) {
	now := time.Now()
	cutoff := now.Add(-chatOnlineTTL)
	if err := s.db.Model(&model.ChatPresence{}).
		Where("is_online = ? AND last_seen_at < ?", true, cutoff).
		Update("is_online", false).Error; err != nil {
		return 0, err
	}

	var count int64
	if err := s.db.Model(&model.ChatPresence{}).
		Where("is_online = ? AND last_seen_at >= ?", true, cutoff).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *ChatService) onlineCountFromCache() int64 {
	v, _ := s.OnlineCount()
	return v
}

func (s *ChatService) touchPresence(userID uint) error {
	now := time.Now()
	cutoff := now.Add(-chatOnlineTTL)
	member := strconv.FormatUint(uint64(userID), 10)
	if score, found, ok := s.cache.ZScore(context.Background(), cacheKeyPresence, member); ok {
		if !found || score < float64(cutoff.UnixMilli()) {
			s.cache.ZRem(context.Background(), cacheKeyPresence, member)
			return errors.New("请先加入聊天室")
		}
		s.markOnline(userID, now)
		return nil
	}

	var presence model.ChatPresence
	if err := s.db.Where("user_id = ?", userID).First(&presence).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("请先加入聊天室")
		}
		return err
	}
	if !presence.IsOnline || presence.LastSeenAt.Before(cutoff) {
		s.db.Model(&model.ChatPresence{}).Where("user_id = ?", userID).Update("is_online", false)
		return errors.New("请先加入聊天室")
	}

	return s.db.Model(&model.ChatPresence{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"is_online":    true,
		"last_seen_at": now,
	}).Error
}

func (s *ChatService) markOnline(userID uint, at time.Time) {
	if s.cache != nil {
		s.cache.ZAdd(context.Background(), cacheKeyPresence, strconv.FormatUint(uint64(userID), 10), float64(at.UnixMilli()))
	}
}

func (s *ChatService) cacheMessage(msg model.ChatMessage) {
	if s.cache == nil {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	ctx := context.Background()
	if !s.cache.ZAdd(ctx, cacheKeyStream, string(data), float64(msg.ID)) {
		return
	}
	s.cache.SetInt(ctx, cacheKeyMaxID, int64(msg.ID), 0)
	s.cache.ZRemRangeByRank(ctx, cacheKeyStream, 0, int64(-chatStreamCap-1))
}

func (s *ChatService) cacheDeletion(messageID uint) {
	if s.cache == nil {
		return
	}
	ctx := context.Background()
	s.cache.ZRemRangeByScore(ctx, cacheKeyStream, float64(messageID), float64(messageID))
	s.cache.ZAdd(ctx, cacheKeyDeleted, strconv.FormatUint(uint64(messageID), 10), float64(nowMs()))
}

func (s *ChatService) getUser(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.Select("id, username, nickname, avatar, status").First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	if user.Status != 1 {
		return nil, errors.New("用户已被禁用")
	}
	return &user, nil
}

func (s *ChatService) fillUserInfo(messages []model.ChatMessage) []model.ChatMessage {
	if len(messages) == 0 {
		return messages
	}

	userIDs := make(map[uint]struct{})
	for _, msg := range messages {
		if msg.UserID > 0 {
			userIDs[msg.UserID] = struct{}{}
		}
	}
	ids := make([]uint, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}

	var users []model.User
	if len(ids) > 0 {
		s.db.Select("id, username, nickname, avatar, real_name, department").Where("id IN ?", ids).Find(&users)
	}
	umap := make(map[uint]model.User, len(users))
	for _, user := range users {
		umap[user.ID] = user
	}

	for i, msg := range messages {
		if user, ok := umap[msg.UserID]; ok {
			messages[i].UserName = displayName(&user)
			messages[i].UserAvatar = user.Avatar
			messages[i].RealName = user.RealName
			messages[i].Department = user.Department
		}
	}
	return messages
}

func displayName(user *model.User) string {
	if user.Nickname != "" {
		return user.Nickname
	}
	return user.Username
}
