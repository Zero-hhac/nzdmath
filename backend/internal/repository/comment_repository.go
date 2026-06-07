package repository

import (
	"math-top/internal/model"

	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *model.Comment) error
	GetByID(id uint) (*model.Comment, error)
	Delete(id uint, userID uint) error
	AdminDelete(id uint) error
	ListByTarget(targetType string, targetID uint, page, pageSize int) ([]model.Comment, int64, error)
	ListReplies(parentID uint, page, pageSize int) ([]model.Comment, int64, error)
	ListWithFilter(filter CommentFilter) ([]model.Comment, int64, error)
	UserHasLiked(userID, commentID uint) bool
	AddLike(userID, commentID uint) error
	RemoveLike(userID, commentID uint) error
	IncrementLikeCount(commentID uint)
	DecrementLikeCount(commentID uint)
	Count() int64
}

type CommentFilter struct {
	Page       int
	PageSize   int
	TargetType string
	Status     *int
}

type commentRepository struct{ db *gorm.DB }

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) GetByID(id uint) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *commentRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Comment{}).Error
}

func (r *commentRepository) AdminDelete(id uint) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

func (r *commentRepository) ListByTarget(targetType string, targetID uint, page, pageSize int) ([]model.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var total int64
	r.db.Model(&model.Comment{}).
		Where("target_type = ? AND target_id = ? AND parent_id IS NULL AND status = 1",
			targetType, targetID).
		Count(&total)
	var comments []model.Comment
	err := r.db.Where("target_type = ? AND target_id = ? AND parent_id IS NULL AND status = 1",
		targetType, targetID).
		Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&comments).Error
	return comments, total, err
}

func (r *commentRepository) ListReplies(parentID uint, page, pageSize int) ([]model.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var total int64
	r.db.Model(&model.Comment{}).Where("parent_id = ? AND status = 1", parentID).Count(&total)
	var comments []model.Comment
	err := r.db.Where("parent_id = ? AND status = 1", parentID).
		Order("id asc").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&comments).Error
	return comments, total, err
}

func (r *commentRepository) ListWithFilter(f CommentFilter) ([]model.Comment, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 10
	}
	tx := r.db.Model(&model.Comment{})
	if f.TargetType != "" {
		tx = tx.Where("target_type = ?", f.TargetType)
	}
	if f.Status != nil {
		tx = tx.Where("status = ?", *f.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var comments []model.Comment
	if err := tx.Order("id desc").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (r *commentRepository) UserHasLiked(userID, commentID uint) bool {
	var n int64
	r.db.Model(&model.CommentLike{}).
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		Count(&n)
	return n > 0
}

func (r *commentRepository) AddLike(userID, commentID uint) error {
	return r.db.Create(&model.CommentLike{
		UserID:    userID,
		CommentID: commentID,
	}).Error
}

func (r *commentRepository) RemoveLike(userID, commentID uint) error {
	return r.db.Where("user_id = ? AND comment_id = ?", userID, commentID).
		Delete(&model.CommentLike{}).Error
}

func (r *commentRepository) IncrementLikeCount(commentID uint) {
	r.db.Model(&model.Comment{}).Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))
}

func (r *commentRepository) DecrementLikeCount(commentID uint) {
	r.db.Model(&model.Comment{}).Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
}

func (r *commentRepository) Count() int64 {
	var n int64
	r.db.Model(&model.Comment{}).Count(&n)
	return n
}