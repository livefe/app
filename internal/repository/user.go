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
	// FindByMobile 根据手机号查找用户
	FindByMobile(mobile string) (*model.User, error)
	// FindByID 根据ID查找用户
	FindByID(id uint) (*model.User, error)
	// Create 创建用户
	Create(user *model.User) error
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

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}
