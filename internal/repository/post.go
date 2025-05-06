package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// PostRepository 动态仓库接口
type PostRepository interface {
	// 动态相关
	CreatePost(post *model.Post) error
	GetPost(id uint) (*model.Post, error)
	GetUserPosts(userID uint, page, size int) ([]model.Post, int64, error)
	GetFollowingPosts(userID uint, page, size int) ([]model.Post, int64, error)
	IncrementPostLikes(postID uint) error
	IncrementPostComments(postID uint) error
}

// postRepository 动态仓库实现
type postRepository struct {
	db *gorm.DB
}

// NewPostRepository 创建动态仓库实例
func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

// CreatePost 创建动态
func (r *postRepository) CreatePost(post *model.Post) error {
	return r.db.Create(post).Error
}

// GetPost 获取动态
func (r *postRepository) GetPost(id uint) (*model.Post, error) {
	var post model.Post
	err := r.db.First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetUserPosts 获取用户动态列表
func (r *postRepository) GetUserPosts(userID uint, page, size int) ([]model.Post, int64, error) {
	var posts []model.Post
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Post{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(size).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

// GetFollowingPosts 获取关注用户的动态列表
func (r *postRepository) GetFollowingPosts(userID uint, page, size int) ([]model.Post, int64, error) {
	var posts []model.Post
	var count int64

	offset := (page - 1) * size

	// 查询关注用户的动态
	query := r.db.Model(&model.Post{}).
		Joins("JOIN relations ON posts.user_id = relations.target_id").
		Where("relations.user_id = ?", userID).
		Where("posts.visibility IN (1, 2)") // 公开或仅好友可见

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("posts.created_at DESC").Offset(offset).Limit(size).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

// IncrementPostLikes 增加动态点赞数
func (r *postRepository) IncrementPostLikes(postID uint) error {
	return r.db.Model(&model.Post{}).Where("id = ?", postID).Update("likes", gorm.Expr("likes + ?", 1)).Error
}

// IncrementPostComments 增加动态评论数
func (r *postRepository) IncrementPostComments(postID uint) error {
	return r.db.Model(&model.Post{}).Where("id = ?", postID).Update("comments", gorm.Expr("comments + ?", 1)).Error
}
