package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// UserFollowerRepository 粉丝关注仓库接口
type UserFollowerRepository interface {
	// 查询方法
	GetFollower(userID, targetID uint) (*model.UserFollower, error)
	GetFollowers(userID uint, page, size int) ([]model.UserFollower, int64, error)
	GetFollowing(userID uint, page, size int) ([]model.UserFollower, int64, error)

	// 修改方法
	CreateFollower(follower *model.UserFollower) error
	DeleteFollower(userID, targetID uint) error
}

// userFollowerRepository 粉丝关注仓库实现
type userFollowerRepository struct {
	db *gorm.DB
}

// NewUserFollowerRepository 创建粉丝关注仓库实例
func NewUserFollowerRepository(db *gorm.DB) UserFollowerRepository {
	return &userFollowerRepository{db: db}
}

// 查询方法

// GetFollower 获取关注关系
func (r *userFollowerRepository) GetFollower(userID, targetID uint) (*model.UserFollower, error) {
	var follower model.UserFollower
	err := r.db.Where("user_id = ? AND target_id = ?", userID, targetID).First(&follower).Error
	if err != nil {
		return nil, err
	}
	return &follower, nil
}

// GetFollowers 获取用户的粉丝列表
func (r *userFollowerRepository) GetFollowers(userID uint, page, size int) ([]model.UserFollower, int64, error) {
	var followers []model.UserFollower
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.UserFollower{}).Where("target_id = ?", userID).Count(&count).Error
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
func (r *userFollowerRepository) GetFollowing(userID uint, page, size int) ([]model.UserFollower, int64, error) {
	var followers []model.UserFollower
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.UserFollower{}).Where("user_id = ?", userID).Count(&count).Error
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
func (r *userFollowerRepository) CreateFollower(follower *model.UserFollower) error {
	return r.db.Create(follower).Error
}

// DeleteFollower 删除关注关系
func (r *userFollowerRepository) DeleteFollower(userID, targetID uint) error {
	return r.db.Where("user_id = ? AND target_id = ?", userID, targetID).Delete(&model.UserFollower{}).Error
}
