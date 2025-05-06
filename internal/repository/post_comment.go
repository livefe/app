package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// PostCommentRepository 动态评论仓库接口
type PostCommentRepository interface {
	// 评论相关
	CreateComment(comment *model.PostComment) error
	GetComment(id uint) (*model.PostComment, error)
	GetPostComments(postID uint, page, size int) ([]model.PostComment, int64, error)
}

// postCommentRepository 动态评论仓库实现
type postCommentRepository struct {
	db *gorm.DB
}

// NewPostCommentRepository 创建动态评论仓库实例
func NewPostCommentRepository(db *gorm.DB) PostCommentRepository {
	return &postCommentRepository{db: db}
}

// CreateComment 创建评论
func (r *postCommentRepository) CreateComment(comment *model.PostComment) error {
	return r.db.Create(comment).Error
}

// GetComment 获取评论详情
func (r *postCommentRepository) GetComment(id uint) (*model.PostComment, error) {
	var comment model.PostComment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetPostComments 获取动态评论列表
func (r *postCommentRepository) GetPostComments(postID uint, page, size int) ([]model.PostComment, int64, error) {
	var comments []model.PostComment
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.PostComment{}).Where("post_id = ?", postID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("post_id = ?", postID).Order("created_at DESC").Offset(offset).Limit(size).Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, count, nil
}
