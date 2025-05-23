package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// PostImageRepository 动态图片存储库接口
type PostImageRepository interface {
	// CreatePostImage 创建动态图片
	CreatePostImage(image *model.PostImage) error
	// GetPostImages 获取动态的所有图片
	GetPostImages(postID uint) ([]model.PostImage, error)
	// DeletePostImage 删除动态图片
	DeletePostImage(id uint) error
	// DeletePostImages 删除动态的所有图片
	DeletePostImages(postID uint) error
	// FindByID 根据ID查找图片
	FindByID(id uint) (*model.PostImage, error)
	// UpdatePostImage 更新图片信息
	UpdatePostImage(image *model.PostImage) error
}

// postImageRepository 动态图片存储库实现
type postImageRepository struct {
	db *gorm.DB
}

// NewPostImageRepository 创建动态图片存储库实例
func NewPostImageRepository(db *gorm.DB) PostImageRepository {
	return &postImageRepository{db: db}
}

// CreatePostImage 创建动态图片
func (r *postImageRepository) CreatePostImage(image *model.PostImage) error {
	return r.db.Create(image).Error
}

// GetPostImages 获取动态的所有图片
func (r *postImageRepository) GetPostImages(postID uint) ([]model.PostImage, error) {
	var images []model.PostImage
	err := r.db.Where("post_id = ?", postID).Find(&images).Error
	return images, err
}

// DeletePostImage 删除动态图片
func (r *postImageRepository) DeletePostImage(id uint) error {
	return r.db.Delete(&model.PostImage{}, id).Error
}

// DeletePostImages 删除动态的所有图片
func (r *postImageRepository) DeletePostImages(postID uint) error {
	return r.db.Where("post_id = ?", postID).Delete(&model.PostImage{}).Error
}

// FindByID 根据ID查找图片
func (r *postImageRepository) FindByID(id uint) (*model.PostImage, error) {
	var image model.PostImage
	err := r.db.First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// UpdatePostImage 更新图片信息
func (r *postImageRepository) UpdatePostImage(image *model.PostImage) error {
	return r.db.Save(image).Error
}
