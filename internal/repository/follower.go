package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// FollowerRepository 粉丝关注仓库接口
type FollowerRepository interface {
	// 查询方法
	GetFollower(userID, targetID uint) (*model.Follower, error)
	GetFollowers(userID uint, page, size int) ([]model.Follower, int64, error)
	GetFollowing(userID uint, page, size int) ([]model.Follower, int64, error)

	// 修改方法
	CreateFollower(follower *model.Follower) error
	DeleteFollower(userID, targetID uint) error
}

// followerRepository 粉丝关注仓库实现
type followerRepository struct {
	db *gorm.DB
}

// NewFollowerRepository 创建粉丝关注仓库实例
func NewFollowerRepository(db *gorm.DB) FollowerRepository {
	return &followerRepository{db: db}
}

// 查询方法

// GetFollower 获取关注关系
func (r *followerRepository) GetFollower(userID, targetID uint) (*model.Follower, error) {
	var follower model.Follower
	err := r.db.Where("user_id = ? AND target_id = ?", userID, targetID).First(&follower).Error
	if err != nil {
		return nil, err
	}
	return &follower, nil
}

// GetFollowers 获取用户的粉丝列表
func (r *followerRepository) GetFollowers(userID uint, page, size int) ([]model.Follower, int64, error) {
	var followers []model.Follower
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Follower{}).Where("target_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("target_id = ?", userID).Offset(offset).Limit(size).Find(&followers).Error
	if err != nil {
		return nil, 0, err
	}

	return followers, count, nil
}

// GetFollowing 获取用户关注的人列表
func (r *followerRepository) GetFollowing(userID uint, page, size int) ([]model.Follower, int64, error) {
	var followers []model.Follower
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Follower{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).Offset(offset).Limit(size).Find(&followers).Error
	if err != nil {
		return nil, 0, err
	}

	return followers, count, nil
}

// 修改方法

// CreateFollower 创建关注关系
func (r *followerRepository) CreateFollower(follower *model.Follower) error {
	return r.db.Create(follower).Error
}

// DeleteFollower 删除关注关系
func (r *followerRepository) DeleteFollower(userID, targetID uint) error {
	return r.db.Where("user_id = ? AND target_id = ?", userID, targetID).Delete(&model.Follower{}).Error
}
