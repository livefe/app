package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// SocialRelationRepository 社交关系仓库接口
type SocialRelationRepository interface {
	// 社交关系相关
	CreateRelation(relation *model.SocialRelation) error
	DeleteRelation(userID, targetID uint) error
	GetRelation(userID, targetID uint) (*model.SocialRelation, error)
	GetFollowers(userID uint, page, size int) ([]model.SocialRelation, int64, error)
	GetFollowing(userID uint, page, size int) ([]model.SocialRelation, int64, error)
}

// socialRelationRepository 社交关系仓库实现
type socialRelationRepository struct {
	db *gorm.DB
}

// NewSocialRelationRepository 创建社交关系仓库实例
func NewSocialRelationRepository(db *gorm.DB) SocialRelationRepository {
	return &socialRelationRepository{db: db}
}

// CreateRelation 创建社交关系
func (r *socialRelationRepository) CreateRelation(relation *model.SocialRelation) error {
	return r.db.Create(relation).Error
}

// DeleteRelation 删除社交关系
func (r *socialRelationRepository) DeleteRelation(userID, targetID uint) error {
	return r.db.Where("user_id = ? AND target_id = ?", userID, targetID).Delete(&model.SocialRelation{}).Error
}

// GetRelation 获取社交关系
func (r *socialRelationRepository) GetRelation(userID, targetID uint) (*model.SocialRelation, error) {
	var relation model.SocialRelation
	err := r.db.Where("user_id = ? AND target_id = ?", userID, targetID).First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// GetFollowers 获取粉丝列表
func (r *socialRelationRepository) GetFollowers(userID uint, page, size int) ([]model.SocialRelation, int64, error) {
	var relations []model.SocialRelation
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.SocialRelation{}).Where("target_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("target_id = ?", userID).Offset(offset).Limit(size).Find(&relations).Error
	if err != nil {
		return nil, 0, err
	}

	return relations, count, nil
}

// GetFollowing 获取关注列表
func (r *socialRelationRepository) GetFollowing(userID uint, page, size int) ([]model.SocialRelation, int64, error) {
	var relations []model.SocialRelation
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.SocialRelation{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).Offset(offset).Limit(size).Find(&relations).Error
	if err != nil {
		return nil, 0, err
	}

	return relations, count, nil
}
