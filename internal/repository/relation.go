package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// RelationRepository 关系仓库接口
type RelationRepository interface {
	// 关系相关
	CreateRelation(relation *model.Relation) error
	DeleteRelation(userID, targetID uint) error
	GetRelation(userID, targetID uint) (*model.Relation, error)
	GetFollowers(userID uint, page, size int) ([]model.Relation, int64, error)
	GetFollowing(userID uint, page, size int) ([]model.Relation, int64, error)
}

// relationRepository 关系仓库实现
type relationRepository struct {
	db *gorm.DB
}

// NewRelationRepository 创建关系仓库实例
func NewRelationRepository(db *gorm.DB) RelationRepository {
	return &relationRepository{db: db}
}

// CreateRelation 创建关系
func (r *relationRepository) CreateRelation(relation *model.Relation) error {
	return r.db.Create(relation).Error
}

// DeleteRelation 删除关系
func (r *relationRepository) DeleteRelation(userID, targetID uint) error {
	return r.db.Where("user_id = ? AND target_id = ?", userID, targetID).Delete(&model.Relation{}).Error
}

// GetRelation 获取关系
func (r *relationRepository) GetRelation(userID, targetID uint) (*model.Relation, error) {
	var relation model.Relation
	err := r.db.Where("user_id = ? AND target_id = ?", userID, targetID).First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// GetFollowers 获取粉丝列表
func (r *relationRepository) GetFollowing(userID uint, page, size int) ([]model.Relation, int64, error) {
	var relations []model.Relation
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Relation{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).Offset(offset).Limit(size).Find(&relations).Error
	if err != nil {
		return nil, 0, err
	}

	return relations, count, nil
}

// GetFollowing 获取关注列表
func (r *relationRepository) GetFollowers(userID uint, page, size int) ([]model.Relation, int64, error) {
	var relations []model.Relation
	var count int64

	offset := (page - 1) * size

	err := r.db.Model(&model.Relation{}).Where("target_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("target_id = ?", userID).Offset(offset).Limit(size).Find(&relations).Error
	if err != nil {
		return nil, 0, err
	}

	return relations, count, nil
}
