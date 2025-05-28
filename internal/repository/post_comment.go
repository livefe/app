package repository

import (
	"app/internal/model"
	"fmt"

	"gorm.io/gorm"
)

// PostCommentRepository 动态评论仓库接口
type PostCommentRepository interface {
	// 评论相关
	CreateComment(comment *model.PostComment) error
	GetComment(id uint) (*model.PostComment, error)
	GetPostComments(postID uint, page, size int) ([]model.PostComment, int64, error)
	// 事务操作
	CreateCommentWithTransaction(comment *model.PostComment, postID uint) error
}

// postCommentRepository 动态评论仓库实现
type postCommentRepository struct {
	db       *gorm.DB
	postRepo PostRepository
}

// NewPostCommentRepository 创建动态评论仓库实例
func NewPostCommentRepository(db *gorm.DB, postRepo PostRepository) PostCommentRepository {
	return &postCommentRepository{db: db, postRepo: postRepo}
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

// CreateCommentWithTransaction 在事务中创建评论并增加评论数
func (r *postCommentRepository) CreateCommentWithTransaction(comment *model.PostComment, postID uint) error {
	// 使用事务确保数据一致性
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 在事务中创建评论
		if err := tx.Create(comment).Error; err != nil {
			return fmt.Errorf("创建评论失败: %w", err)
		}

		// 在事务中增加评论数，使用postRepo的事务方法
		if err := r.postRepo.IncrementPostCommentsWithTx(tx, postID); err != nil {
			return fmt.Errorf("增加评论数失败: %w", err)
		}

		return nil
	})
}
