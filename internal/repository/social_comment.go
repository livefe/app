package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// SocialCommentRepository 社交评论仓库接口
type SocialCommentRepository interface {
	// 评论相关
	CreateComment(comment *model.SocialComment) error
	GetComment(id uint) (*model.SocialComment, error)
	GetPostComments(postID uint, page, size int) ([]model.SocialComment, int64, error)
}

// socialCommentRepository 社交评论仓库实现
type socialCommentRepository struct {
	db *gorm.DB
}

// NewSocialCommentRepository 创建社交评论仓库实例
func NewSocialCommentRepository(db *gorm.DB) SocialCommentRepository {
	return &socialCommentRepository{db: db}
}

// CreateComment 创建评论
func (r *socialCommentRepository) CreateComment(comment *model.SocialComment) error {
	return r.db.Create(comment).Error
}

// GetComment 获取评论详情
func (r *socialCommentRepository) GetComment(id uint) (*model.SocialComment, error) {
	var comment model.SocialComment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetPostComments 获取动态评论列表
func (r *socialCommentRepository) GetPostComments(postID uint, page, size int) ([]model.SocialComment, int64, error) {
	var comments []model.SocialComment
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.SocialComment{}).Where("post_id = ?", postID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("post_id = ?", postID).Order("created_at DESC").Offset(offset).Limit(size).Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, count, nil
}
