package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// FriendRepository 好友关系仓库接口
type FriendRepository interface {
	// 好友相关
	CreateFriend(friend *model.Friend) error
	UpdateFriendStatus(id uint, status int) error
	DeleteFriend(userID, targetID uint) error
	GetFriend(userID, targetID uint) (*model.Friend, error)
	GetFriendByID(id uint) (*model.Friend, error)
	GetFriendRequests(userID uint, page, size int) ([]model.Friend, int64, error)
	GetFriends(userID uint, page, size int) ([]model.Friend, int64, error)
}

// friendRepository 好友关系仓库实现
type friendRepository struct {
	db *gorm.DB
}

// NewFriendRepository 创建好友关系仓库实例
func NewFriendRepository(db *gorm.DB) FriendRepository {
	return &friendRepository{db: db}
}

// CreateFriend 创建好友关系
func (r *friendRepository) CreateFriend(friend *model.Friend) error {
	return r.db.Create(friend).Error
}

// UpdateFriendStatus 更新好友关系状态
func (r *friendRepository) UpdateFriendStatus(id uint, status int) error {
	return r.db.Model(&model.Friend{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteFriend 删除好友关系
func (r *friendRepository) DeleteFriend(userID, targetID uint) error {
	return r.db.Where(
		"(user_id = ? AND target_id = ?) OR (user_id = ? AND target_id = ?)",
		userID, targetID, targetID, userID,
	).Delete(&model.Friend{}).Error
}

// GetFriend 获取好友关系
func (r *friendRepository) GetFriend(userID, targetID uint) (*model.Friend, error) {
	var friend model.Friend
	err := r.db.Where(
		"(user_id = ? AND target_id = ?) OR (user_id = ? AND target_id = ?)",
		userID, targetID, targetID, userID,
	).First(&friend).Error
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

// GetFriendByID 根据ID获取好友关系
func (r *friendRepository) GetFriendByID(id uint) (*model.Friend, error) {
	var friend model.Friend
	err := r.db.Where("id = ?", id).First(&friend).Error
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

// GetFriendRequests 获取好友请求列表
func (r *friendRepository) GetFriendRequests(userID uint, page, size int) ([]model.Friend, int64, error) {
	var friends []model.Friend
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Friend{}).Where("target_id = ? AND status = 0", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("target_id = ? AND status = 0", userID).Offset(offset).Limit(size).Find(&friends).Error
	if err != nil {
		return nil, 0, err
	}

	return friends, count, nil
}

// GetFriends 获取好友列表
func (r *friendRepository) GetFriends(userID uint, page, size int) ([]model.Friend, int64, error) {
	var friends []model.Friend
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Friend{}).Where(
		"(user_id = ? OR target_id = ?) AND status = 1",
		userID, userID,
	).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where(
		"(user_id = ? OR target_id = ?) AND status = 1",
		userID, userID,
	).Offset(offset).Limit(size).Find(&friends).Error
	if err != nil {
		return nil, 0, err
	}

	return friends, count, nil
}
