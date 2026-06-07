package service

import (
	"errors"
	"math-top/internal/model"
	"strconv"

	"gorm.io/gorm"
)

type CommentService struct {
	db *gorm.DB
}

func NewCommentService(db *gorm.DB) *CommentService {
	return &CommentService{db: db}
}

type CommentWithReplies struct {
	model.Comment
	Replies []model.Comment `json:"replies"`
}

type CreateCommentParams struct {
	TargetType string
	TargetID   uint
	Content    string
	Rating     int
	ParentID   *uint
}

var validTargetTypes = map[string]bool{
	"event": true, "news": true, "resource": true, "showcase": true,
}

func (s *CommentService) Create(userID uint, p CreateCommentParams) (*model.Comment, error) {
	if p.Content == "" {
		return nil, errors.New("评论内容不能为空")
	}
	if !validTargetTypes[p.TargetType] {
		return nil, errors.New("不支持的评论目标类型")
	}
	if p.Rating < 0 || p.Rating > 5 {
		return nil, errors.New("评分需在 0-5 之间")
	}
	if p.ParentID != nil {
		var parent model.Comment
		if err := s.db.First(&parent, *p.ParentID).Error; err != nil {
			return nil, errors.New("父评论不存在")
		}
		if parent.ParentID != nil {
			return nil, errors.New("暂不支持多层回复")
		}
	}
	comment := &model.Comment{
		UserID:     userID,
		TargetType: p.TargetType,
		TargetID:   p.TargetID,
		Content:    p.Content,
		Rating:     p.Rating,
		ParentID:   p.ParentID,
		Status:     1,
	}
	if err := s.db.Create(comment).Error; err != nil {
		return nil, errors.New("发表评论失败")
	}
	if p.ParentID != nil {
		s.db.Model(&model.Comment{}).Where("id = ?", *p.ParentID).
			UpdateColumn("reply_count", gorm.Expr("reply_count + 1"))
	}
	return comment, nil
}

func (s *CommentService) ListByTarget(targetType string, targetID uint, page, pageSize int) ([]CommentWithReplies, int64, error) {
	if !validTargetTypes[targetType] {
		return nil, 0, errors.New("不支持的评论目标类型")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var total int64
	s.db.Model(&model.Comment{}).
		Where("target_type = ? AND target_id = ? AND parent_id IS NULL AND status = 1", targetType, targetID).
		Count(&total)

	var comments []model.Comment
	if err := s.db.Where("target_type = ? AND target_id = ? AND parent_id IS NULL AND status = 1", targetType, targetID).
		Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	result := make([]CommentWithReplies, 0, len(comments))
	for _, c := range comments {
		var replies []model.Comment
		s.db.Where("parent_id = ? AND status = 1", c.ID).
			Order("id asc").Limit(50).
			Find(&replies)
		result = append(result, CommentWithReplies{Comment: c, Replies: replies})
	}
	return result, total, nil
}

func (s *CommentService) Delete(userID, commentID uint) error {
	res := s.db.Where("id = ? AND user_id = ?", commentID, userID).Delete(&model.Comment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("评论不存在或无权限")
	}
	return nil
}

func (s *CommentService) ToggleLike(userID, commentID uint) (bool, error) {
	var comment model.Comment
	if err := s.db.First(&comment, commentID).Error; err != nil {
		return false, errors.New("评论不存在")
	}
	var existing model.CommentLike
	err := s.db.Where("user_id = ? AND comment_id = ?", userID, commentID).First(&existing).Error
	if err == nil {
		s.db.Where("user_id = ? AND comment_id = ?", userID, commentID).Delete(&model.CommentLike{})
		s.db.Model(&model.Comment{}).Where("id = ?", commentID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
		return false, nil
	}
	s.db.Create(&model.CommentLike{UserID: userID, CommentID: commentID})
	s.db.Model(&model.Comment{}).Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	return true, nil
}

func (s *CommentService) ListWithFilter(page, pageSize int, targetType string, status *int) ([]model.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	tx := s.db.Model(&model.Comment{})
	if targetType != "" {
		tx = tx.Where("target_type = ?", targetType)
	}
	if status != nil {
		tx = tx.Where("status = ?", *status)
	}
	var total int64
	tx.Count(&total)
	var comments []model.Comment
	if err := tx.Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (s *CommentService) AdminDelete(commentID uint) error {
	res := s.db.Delete(&model.Comment{}, commentID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("评论不存在")
	}
	return nil
}

func (s *CommentService) SetStatus(commentID uint, status int) error {
	res := s.db.Model(&model.Comment{}).Where("id = ?", commentID).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("评论不存在")
	}
	return nil
}

func (s *CommentService) FillUserInfo(comments []model.Comment) []model.Comment {
	if len(comments) == 0 {
		return comments
	}
	userIDs := make([]uint, 0, len(comments))
	for _, c := range comments {
		userIDs = append(userIDs, c.UserID)
	}
	var users []model.User
	s.db.Select("id, username, nickname, avatar").Where("id IN ?", userIDs).Find(&users)
	umap := make(map[uint]model.User, len(users))
	for _, u := range users {
		umap[u.ID] = u
	}
	for i, c := range comments {
		if u, ok := umap[c.UserID]; ok {
			comments[i].UserName = u.Nickname
			if comments[i].UserName == "" {
				comments[i].UserName = u.Username
			}
			comments[i].UserAvatar = u.Avatar
		}
	}
	return comments
}

func (s *CommentService) FillUserInfoAsSlice(comments []CommentWithReplies) []CommentWithReplies {
	if len(comments) == 0 {
		return comments
	}
	userIDs := make(map[uint]struct{})
	for _, c := range comments {
		userIDs[c.UserID] = struct{}{}
		for _, r := range c.Replies {
			userIDs[r.UserID] = struct{}{}
		}
	}
	ids := make([]uint, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}
	var users []model.User
	s.db.Select("id, username, nickname, avatar").Where("id IN ?", ids).Find(&users)
	umap := make(map[uint]model.User, len(users))
	for _, u := range users {
		umap[u.ID] = u
	}
	for i, c := range comments {
		if u, ok := umap[c.UserID]; ok {
			name := u.Nickname
			if name == "" {
				name = u.Username
			}
			comments[i].UserName = name
			comments[i].UserAvatar = u.Avatar
		}
		for j, r := range c.Replies {
			if u, ok := umap[r.UserID]; ok {
				name := u.Nickname
				if name == "" {
					name = u.Username
				}
				comments[i].Replies[j].UserName = name
				comments[i].Replies[j].UserAvatar = u.Avatar
			}
		}
	}
	return comments
}

var _ = strconv.Itoa