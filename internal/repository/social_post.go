package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// SocialPostRepository 社交动态仓库接口
type SocialPostRepository interface {
	// 社交动态相关
	CreatePost(post *model.SocialPost) error
	GetPost(id uint) (*model.SocialPost, error)
	GetUserPosts(userID uint, page, size int) ([]model.SocialPost, int64, error)
	GetFollowingPosts(userID uint, page, size int) ([]model.SocialPost, int64, error)
	IncrementPostLikes(postID uint) error
	IncrementPostComments(postID uint) error
}

// socialPostRepository 社交动态仓库实现
type socialPostRepository struct {
	db *gorm.DB
}

// NewSocialPostRepository 创建社交动态仓库实例
func NewSocialPostRepository(db *gorm.DB) SocialPostRepository {
	return &socialPostRepository{db: db}
}

// CreatePost 创建社交动态
func (r *socialPostRepository) CreatePost(post *model.SocialPost) error {
	return r.db.Create(post).Error
}

// GetPost 获取社交动态
func (r *socialPostRepository) GetPost(id uint) (*model.SocialPost, error) {
	var post model.SocialPost
	err := r.db.First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetUserPosts 获取用户社交动态列表
func (r *socialPostRepository) GetUserPosts(userID uint, page, size int) ([]model.SocialPost, int64, error) {
	var posts []model.SocialPost
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.SocialPost{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(size).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

// GetFollowingPosts 获取关注用户的社交动态列表
func (r *socialPostRepository) GetFollowingPosts(userID uint, page, size int) ([]model.SocialPost, int64, error) {
	var posts []model.SocialPost
	var count int64

	offset := (page - 1) * size

	// 查询关注用户的动态
	query := r.db.Model(&model.SocialPost{}).
		Joins("JOIN social_relations ON social_posts.user_id = social_relations.target_id").
		Where("social_relations.user_id = ?", userID).
		Where("social_posts.visibility IN (1, 2)") // 公开或仅好友可见

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("social_posts.created_at DESC").Offset(offset).Limit(size).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

// IncrementPostLikes 增加动态点赞数
func (r *socialPostRepository) IncrementPostLikes(postID uint) error {
	return r.db.Model(&model.SocialPost{}).Where("id = ?", postID).Update("likes", gorm.Expr("likes + ?", 1)).Error
}

// IncrementPostComments 增加动态评论数
func (r *socialPostRepository) IncrementPostComments(postID uint) error {
	return r.db.Model(&model.SocialPost{}).Where("id = ?", postID).Update("comments", gorm.Expr("comments + ?", 1)).Error
}
