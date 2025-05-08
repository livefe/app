package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// UserFriendRepository 好友关系仓库接口
type UserFriendRepository interface {
	// 好友相关
	CreateFriend(friend *model.UserFriend) error
	UpdateFriendStatus(id uint, status int) error
	DeleteFriend(userID, targetID uint) error
	GetFriend(userID, targetID uint) (*model.UserFriend, error)
	GetFriendByID(id uint) (*model.UserFriend, error)
	GetFriendRequests(userID uint, page, size int) ([]model.UserFriend, int64, error)
	GetFriends(userID uint, page, size int) ([]model.UserFriend, int64, error)
}

// userFriendRepository 好友关系仓库实现
type userFriendRepository struct {
	db *gorm.DB
}

// NewUserFriendRepository 创建好友关系仓库实例
func NewUserFriendRepository(db *gorm.DB) UserFriendRepository {
	return &userFriendRepository{db: db}
}

// CreateFriend 创建好友关系（双记录模式）
func (r *userFriendRepository) CreateFriend(friend *model.UserFriend) error {
	// 开启事务
	tx := r.db.Begin()

	// 创建发起方记录
	friend.Direction = 0 // 发起方
	if err := tx.Create(friend).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建接收方视角的记录
	receiverFriend := &model.UserFriend{
		UserID:    friend.TargetID, // 接收方视角：自己是UserID
		TargetID:  friend.UserID,   // 接收方视角：对方是TargetID
		Status:    friend.Status,
		Direction: 1, // 接收方
	}

	if err := tx.Create(receiverFriend).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// UpdateFriendStatus 更新好友关系状态（双记录模式）
func (r *userFriendRepository) UpdateFriendStatus(id uint, status int) error {
	// 先查询要更新的记录，获取UserID和TargetID
	var friend model.UserFriend
	if err := r.db.Where("id = ?", id).First(&friend).Error; err != nil {
		return err
	}

	// 开启事务
	tx := r.db.Begin()

	// 更新当前记录状态
	if err := tx.Model(&model.UserFriend{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新对应的另一条记录状态
	if err := tx.Model(&model.UserFriend{}).Where(
		"user_id = ? AND target_id = ?",
		friend.TargetID, friend.UserID,
	).Update("status", status).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// DeleteFriend 删除好友关系（双记录模式）
func (r *userFriendRepository) DeleteFriend(userID, targetID uint) error {
	// 开启事务
	tx := r.db.Begin()

	// 删除第一条记录（用户视角）
	if err := tx.Where("user_id = ? AND target_id = ?", userID, targetID).Delete(&model.UserFriend{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除第二条记录（好友视角）
	if err := tx.Where("user_id = ? AND target_id = ?", targetID, userID).Delete(&model.UserFriend{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// GetFriend 获取好友关系（双记录模式）
func (r *userFriendRepository) GetFriend(userID, targetID uint) (*model.UserFriend, error) {
	// 在双记录模式下，只需要查询用户视角的记录
	var friend model.UserFriend
	err := r.db.Where("user_id = ? AND target_id = ?", userID, targetID).First(&friend).Error
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

// GetFriendByID 根据ID获取好友关系
func (r *userFriendRepository) GetFriendByID(id uint) (*model.UserFriend, error) {
	var friend model.UserFriend
	err := r.db.Where("id = ?", id).First(&friend).Error
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

// GetFriendRequests 获取好友请求列表（双记录模式）
func (r *userFriendRepository) GetFriendRequests(userID uint, page, size int) ([]model.UserFriend, int64, error) {
	var friends []model.UserFriend
	var count int64

	offset := (page - 1) * size

	// 在双记录模式下，查询用户视角下的待确认请求
	// 用户是接收方(Direction=1)且状态为待确认(Status=0)
	err := r.db.Model(&model.UserFriend{}).Where(
		"user_id = ? AND status = 0 AND direction = 1",
		userID,
	).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where(
		"user_id = ? AND status = 0 AND direction = 1",
		userID,
	).Offset(offset).Limit(size).Find(&friends).Error
	if err != nil {
		return nil, 0, err
	}

	return friends, count, nil
}

// GetFriends 获取好友列表（双记录模式）
func (r *userFriendRepository) GetFriends(userID uint, page, size int) ([]model.UserFriend, int64, error) {
	var friends []model.UserFriend
	var count int64

	offset := (page - 1) * size

	// 在双记录模式下，只需要查询用户视角下的已确认好友
	// 用户是记录所有者(UserID=userID)且状态为已确认(Status=1)
	err := r.db.Model(&model.UserFriend{}).Where(
		"user_id = ? AND status = 1",
		userID,
	).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where(
		"user_id = ? AND status = 1",
		userID,
	).Offset(offset).Limit(size).Find(&friends).Error
	if err != nil {
		return nil, 0, err
	}

	return friends, count, nil
}