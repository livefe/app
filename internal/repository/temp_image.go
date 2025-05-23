package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// TempImageRepository 临时图片存储库接口
type TempImageRepository interface {
	// CreateTempImage 创建临时图片
	CreateTempImage(image *model.TempImage) error
	// FindByID 根据ID查找临时图片
	FindByID(id uint) (*model.TempImage, error)
	// UpdateTempImage 更新临时图片信息
	UpdateTempImage(image *model.TempImage) error
	// DeleteTempImage 删除临时图片
	DeleteTempImage(id uint) error
	// GetUserTempImages 获取用户的所有临时图片
	GetUserTempImages(userID uint) ([]model.TempImage, error)
}

// tempImageRepository 临时图片存储库实现
type tempImageRepository struct {
	db *gorm.DB
}

// NewTempImageRepository 创建临时图片存储库实例
func NewTempImageRepository(db *gorm.DB) TempImageRepository {
	return &tempImageRepository{db: db}
}

// CreateTempImage 创建临时图片
func (r *tempImageRepository) CreateTempImage(image *model.TempImage) error {
	return r.db.Create(image).Error
}

// FindByID 根据ID查找临时图片
func (r *tempImageRepository) FindByID(id uint) (*model.TempImage, error) {
	var image model.TempImage
	err := r.db.First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// UpdateTempImage 更新临时图片信息
func (r *tempImageRepository) UpdateTempImage(image *model.TempImage) error {
	return r.db.Save(image).Error
}

// DeleteTempImage 删除临时图片
func (r *tempImageRepository) DeleteTempImage(id uint) error {
	return r.db.Delete(&model.TempImage{}, id).Error
}

// GetUserTempImages 获取用户的所有临时图片
func (r *tempImageRepository) GetUserTempImages(userID uint) ([]model.TempImage, error) {
	var images []model.TempImage
	err := r.db.Where("user_id = ?", userID).Find(&images).Error
	return images, err
}
