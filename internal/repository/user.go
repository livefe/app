package repository

import (
	"errors"

	"app/internal/model"

	"gorm.io/gorm"
)

var (
	// ErrRecordNotFound 记录未找到错误
	ErrRecordNotFound = errors.New("记录未找到")
)

// UserRepository 用户仓库接口
type UserRepository interface {
	// 查询方法
	FindByID(id uint) (*model.User, error)           // 根据ID查找用户
	FindByMobile(mobile string) (*model.User, error) // 根据手机号查找用户

	// 修改方法
	Create(user *model.User) error // 创建用户
	SoftDelete(id uint) error      // 软删除用户（注销账号）
}

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// FindByID 根据ID查找用户
func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByMobile 根据手机号查找用户
func (r *userRepository) FindByMobile(mobile string) (*model.User, error) {
	var user model.User
	result := r.db.Where("mobile = ?", mobile).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// SoftDelete 软删除用户（注销账号）
func (r *userRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&model.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
