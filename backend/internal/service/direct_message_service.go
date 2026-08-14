package service

import (
	"errors"
	"fmt"
	"io"
	"math-top/internal/config"
	"math-top/internal/dto"
	"math-top/internal/model"
	"math-top/internal/ws"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DirectMessageService struct {
	db  *gorm.DB
	hub *ws.Hub
}

func NewDirectMessageService(db *gorm.DB, hub *ws.Hub) *DirectMessageService {
	return &DirectMessageService{
		db:  db,
		hub: hub,
	}
}

// GetMyMessages 会员获取自己与管理员的私聊消息列表（带基于 beforeID 的分页）
func (s *DirectMessageService) GetMyMessages(userID uint, beforeID uint, limit int) ([]dto.DirectMessageItem, bool, uint, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}

	query := s.db.Where("user_id = ?", userID).Order("id desc")
	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}

	var messages []model.DirectMessage
	if err := query.Limit(limit + 1).Find(&messages).Error; err != nil {
		return nil, false, 0, errors.New("获取私聊记录失败")
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	// 逆序转为按时间正序
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	var nextBeforeID uint
	if len(messages) > 0 {
		nextBeforeID = messages[0].ID
	}

	return s.enrichMessages(messages), hasMore, nextBeforeID, nil
}

// SendUserMessage 会员向管理员发送文本消息
func (s *DirectMessageService) SendUserMessage(userID uint, content string) (*dto.DirectMessageItem, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("消息内容不能为空")
	}
	if len([]rune(content)) > 2000 {
		return nil, errors.New("消息内容不能超过 2000 字")
	}

	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	msg := &model.DirectMessage{
		UserID:      userID,
		SenderType:  "user",
		MessageType: "text",
		Content:     content,
		IsRead:      false,
	}
	if err := s.db.Create(msg).Error; err != nil {
		return nil, errors.New("发送私聊消息失败")
	}

	items := s.enrichMessages([]model.DirectMessage{*msg})
	res := &items[0]

	// 实时推送：只投递到发送方与管理员会话侧，绝不全员广播（私信机密性）
	s.pushDirect(append(s.adminUserIDs(), userID), res)

	return res, nil
}

// SendUserFile 会员向管理员发送文件/图片
func (s *DirectMessageService) SendUserFile(userID uint, file *multipart.FileHeader) (*dto.DirectMessageItem, error) {
	if file == nil {
		return nil, errors.New("请上传文件")
	}
	if file.Size > 5*1024*1024 {
		return nil, errors.New("文件不能超过 5MB")
	}

	originalName := filepath.Base(file.Filename)
	ext := strings.ToLower(filepath.Ext(originalName))
	mime := file.Header.Get("Content-Type")
	if chatVideoExts[ext] || strings.HasPrefix(strings.ToLower(mime), "video/") {
		return nil, errors.New("不支持上传视频文件")
	}
	if !allowedChatExt(ext) {
		return nil, errors.New("不支持的文件格式")
	}

	msgType := "file"
	if chatImageExts[ext] {
		msgType = "image"
	}

	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}

	now := time.Now()
	dir := filepath.Join(uploadBase, "direct", now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(dir, 0775); err != nil {
		return nil, errors.New("创建上传目录失败")
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

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(savePath)
		return nil, errors.New("写入文件失败")
	}

	fileURL := fmt.Sprintf("/uploads/direct/%s/%s/%s", now.Format("2006"), now.Format("01"), saveName)

	msg := &model.DirectMessage{
		UserID:      userID,
		SenderType:  "user",
		MessageType: msgType,
		FileName:    originalName,
		FileURL:     fileURL,
		FilePath:    savePath,
		FileSize:    file.Size,
		FileExt:     ext,
		FileMime:    mime,
		IsRead:      false,
	}
	if err := s.db.Create(msg).Error; err != nil {
		os.Remove(savePath)
		return nil, errors.New("发送文件失败")
	}

	items := s.enrichMessages([]model.DirectMessage{*msg})
	res := &items[0]

	s.pushDirect(append(s.adminUserIDs(), userID), res)

	return res, nil
}

// MarkUserRead 会员标记管理员消息为已读
func (s *DirectMessageService) MarkUserRead(userID uint) error {
	return s.db.Model(&model.DirectMessage{}).
		Where("user_id = ? AND sender_type = 'admin' AND is_read = false", userID).
		Update("is_read", true).Error
}

// GetUnreadCountForUser 会员获取未读管理员消息数
func (s *DirectMessageService) GetUnreadCountForUser(userID uint) (int64, error) {
	var count int64
	err := s.db.Model(&model.DirectMessage{}).
		Where("user_id = ? AND sender_type = 'admin' AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

// ListConversations 管理员获取所有会话列表
func (s *DirectMessageService) ListConversations(keyword string, page, pageSize int) ([]dto.DirectConversationItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 找出所有有私聊记录的用户 ID 及其最后一条消息时间
	type UserChatSummary struct {
		UserID        uint      `gorm:"column:user_id"`
		LastMessageAt time.Time `gorm:"column:last_message_at"`
	}

	var summaries []UserChatSummary
	subQuery := s.db.Model(&model.DirectMessage{}).
		Select("user_id, MAX(created_at) as last_message_at").
		Group("user_id")

	var total int64
	if err := s.db.Table("(?) as ucs", subQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Table("(?) as ucs", subQuery).
		Order("last_message_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&summaries).Error; err != nil {
		return nil, 0, err
	}

	if len(summaries) == 0 {
		return []dto.DirectConversationItem{}, total, nil
	}

	userIDs := make([]uint, 0, len(summaries))
	for _, s := range summaries {
		userIDs = append(userIDs, s.UserID)
	}

	// 批量查询用户信息
	var users []model.User
	userQuery := s.db.Where("id IN ?", userIDs)
	if keyword != "" {
		userQuery = userQuery.Where("username LIKE ? OR nickname LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	userQuery.Find(&users)
	userMap := make(map[uint]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	result := make([]dto.DirectConversationItem, 0, len(summaries))
	for _, sum := range summaries {
		u, ok := userMap[sum.UserID]
		if !ok && keyword != "" {
			continue
		}

		// 查最后一条消息
		var lastMsg model.DirectMessage
		s.db.Where("user_id = ?", sum.UserID).Order("id desc").First(&lastMsg)

		// 查该用户发给管理员的未读消息数
		var unread int64
		s.db.Model(&model.DirectMessage{}).
			Where("user_id = ? AND sender_type = 'user' AND is_read = false", sum.UserID).
			Count(&unread)

		item := dto.DirectConversationItem{
			UserID:          sum.UserID,
			Username:        u.Username,
			Nickname:        u.Nickname,
			RealName:        u.RealName,
			ClassName:       u.ClassName,
			Department:      u.Department,
			Avatar:          u.Avatar,
			LastMessage:     lastMsg.Content,
			LastMessageType: lastMsg.MessageType,
			LastMessageAt:   lastMsg.CreatedAt,
			UnreadCount:     unread,
		}
		if item.LastMessage == "" && lastMsg.FileName != "" {
			item.LastMessage = "[" + lastMsg.FileName + "]"
		}
		result = append(result, item)
	}

	return result, total, nil
}

// GetMessagesByUser 管理员获取指定用户的完整私聊记录
func (s *DirectMessageService) GetMessagesByUser(targetUserID uint, beforeID uint, limit int) ([]dto.DirectMessageItem, bool, uint, error) {
	return s.GetMyMessages(targetUserID, beforeID, limit)
}

// SendAdminMessage 管理员向指定用户发送文本回复
func (s *DirectMessageService) SendAdminMessage(adminID uint, targetUserID uint, content string) (*dto.DirectMessageItem, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("消息内容不能为空")
	}
	if len([]rune(content)) > 2000 {
		return nil, errors.New("消息内容不能超过 2000 字")
	}

	var user model.User
	if err := s.db.First(&user, targetUserID).Error; err != nil {
		return nil, errors.New("目标用户不存在")
	}

	msg := &model.DirectMessage{
		UserID:      targetUserID,
		SenderType:  "admin",
		AdminID:     adminID,
		MessageType: "text",
		Content:     content,
		IsRead:      false,
	}
	if err := s.db.Create(msg).Error; err != nil {
		return nil, errors.New("发送回复失败")
	}

	items := s.enrichMessages([]model.DirectMessage{*msg})
	res := &items[0]

	// 定向推送：仅管理员发送方与目标会员可见
	s.pushDirect([]uint{adminID, targetUserID}, res)

	return res, nil
}

// SendAdminFile 管理员向指定用户发送文件/图片
func (s *DirectMessageService) SendAdminFile(adminID uint, targetUserID uint, file *multipart.FileHeader) (*dto.DirectMessageItem, error) {
	if file == nil {
		return nil, errors.New("请上传文件")
	}
	if file.Size > 5*1024*1024 {
		return nil, errors.New("文件不能超过 5MB")
	}

	originalName := filepath.Base(file.Filename)
	ext := strings.ToLower(filepath.Ext(originalName))
	mime := file.Header.Get("Content-Type")
	if chatVideoExts[ext] || strings.HasPrefix(strings.ToLower(mime), "video/") {
		return nil, errors.New("不支持上传视频文件")
	}
	if !allowedChatExt(ext) {
		return nil, errors.New("不支持的文件格式")
	}

	msgType := "file"
	if chatImageExts[ext] {
		msgType = "image"
	}

	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}

	now := time.Now()
	dir := filepath.Join(uploadBase, "direct", now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(dir, 0775); err != nil {
		return nil, errors.New("创建上传目录失败")
	}

	saveName := fmt.Sprintf("adm%d_u%d_%d%s", adminID, targetUserID, now.UnixNano(), ext)
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

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(savePath)
		return nil, errors.New("写入文件失败")
	}

	fileURL := fmt.Sprintf("/uploads/direct/%s/%s/%s", now.Format("2006"), now.Format("01"), saveName)

	msg := &model.DirectMessage{
		UserID:      targetUserID,
		SenderType:  "admin",
		AdminID:     adminID,
		MessageType: msgType,
		FileName:    originalName,
		FileURL:     fileURL,
		FilePath:    savePath,
		FileSize:    file.Size,
		FileExt:     ext,
		FileMime:    mime,
		IsRead:      false,
	}
	if err := s.db.Create(msg).Error; err != nil {
		os.Remove(savePath)
		return nil, errors.New("发送文件失败")
	}

	items := s.enrichMessages([]model.DirectMessage{*msg})
	res := &items[0]

	s.pushDirect([]uint{adminID, targetUserID}, res)

	return res, nil
}

// MarkAdminRead 管理员标记指定用户的消息为已读
func (s *DirectMessageService) MarkAdminRead(targetUserID uint) error {
	return s.db.Model(&model.DirectMessage{}).
		Where("user_id = ? AND sender_type = 'user' AND is_read = false", targetUserID).
		Update("is_read", true).Error
}

// pushDirect 向指定用户集合定向推送私信帧（channel=direct 标记保持不变，前端无需改动）。
// 私信只允许会话双方（及管理员侧）收到，禁止全员广播。
func (s *DirectMessageService) pushDirect(userIDs []uint, res *dto.DirectMessageItem) {
	if s.hub == nil || len(userIDs) == 0 {
		return
	}
	s.hub.SendToUsers(userIDs, map[string]interface{}{
		"channel": "direct",
		"message": res,
	})
}

// adminUserIDs 返回当前管理员账号对应的会员表用户 ID（管理员登录后会自动同步会员记录）。
func (s *DirectMessageService) adminUserIDs() []uint {
	var users []model.User
	if err := s.db.Select("id").Where("role IN ?", []string{"admin", "super_admin"}).Find(&users).Error; err != nil {
		return nil
	}
	ids := make([]uint, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids
}

// GetTotalUnreadForAdmin 获取管理员全局未读私信数
func (s *DirectMessageService) GetTotalUnreadForAdmin() (int64, error) {
	var count int64
	err := s.db.Model(&model.DirectMessage{}).
		Where("sender_type = 'user' AND is_read = false").
		Count(&count).Error
	return count, err
}

func (s *DirectMessageService) enrichMessages(messages []model.DirectMessage) []dto.DirectMessageItem {
	if len(messages) == 0 {
		return []dto.DirectMessageItem{}
	}

	userIDs := make([]uint, 0)
	for _, m := range messages {
		userIDs = append(userIDs, m.UserID)
	}

	var users []model.User
	s.db.Select("id, username, nickname, real_name, avatar").Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[uint]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	out := make([]dto.DirectMessageItem, 0, len(messages))
	for _, m := range messages {
		item := dto.DirectMessageItem{
			ID:          m.ID,
			UserID:      m.UserID,
			SenderType:  m.SenderType,
			AdminID:     m.AdminID,
			MessageType: m.MessageType,
			Content:     m.Content,
			FileName:    m.FileName,
			FileURL:     m.FileURL,
			FileSize:    m.FileSize,
			FileExt:     m.FileExt,
			IsRead:      m.IsRead,
			CreatedAt:   m.CreatedAt,
		}
		if m.SenderType == "admin" {
			item.SenderName = "协会管理员"
		} else if u, ok := userMap[m.UserID]; ok {
			if u.Nickname != "" {
				item.SenderName = u.Nickname
			} else if u.RealName != "" {
				item.SenderName = u.RealName
			} else {
				item.SenderName = u.Username
			}
			item.SenderAvatar = u.Avatar
		} else {
			item.SenderName = fmt.Sprintf("用户 #%d", m.UserID)
		}
		out = append(out, item)
	}
	return out
}
