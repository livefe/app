package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// SocialLocationShareRepository 位置分享仓库接口
type SocialLocationShareRepository interface {
	// 位置分享相关
	CreateLocationShare(location *model.SocialLocationShare) error
	GetLocationShare(id uint) (*model.SocialLocationShare, error)
	GetUserLocationShare(userID uint) (*model.SocialLocationShare, error)
	GetNearbyUsers(userID uint, lat, lng, radius float64, page, size int) ([]model.SocialLocationShare, int64, error)
}

// socialLocationShareRepository 位置分享仓库实现
type socialLocationShareRepository struct {
	db *gorm.DB
}

// NewSocialLocationShareRepository 创建位置分享仓库实例
func NewSocialLocationShareRepository(db *gorm.DB) SocialLocationShareRepository {
	return &socialLocationShareRepository{db: db}
}

// CreateLocationShare 创建位置分享
func (r *socialLocationShareRepository) CreateLocationShare(location *model.SocialLocationShare) error {
	return r.db.Create(location).Error
}

// GetLocationShare 获取位置分享
func (r *socialLocationShareRepository) GetLocationShare(id uint) (*model.SocialLocationShare, error) {
	var location model.SocialLocationShare
	err := r.db.First(&location, id).Error
	if err != nil {
		return nil, err
	}
	return &location, nil
}

// GetUserLocationShare 获取用户最新位置分享
func (r *socialLocationShareRepository) GetUserLocationShare(userID uint) (*model.SocialLocationShare, error) {
	var location model.SocialLocationShare
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&location).Error
	if err != nil {
		return nil, err
	}
	return &location, nil
}

// GetNearbyUsers 获取附近用户
func (r *socialLocationShareRepository) GetNearbyUsers(userID uint, lat, lng, radius float64, page, size int) ([]model.SocialLocationShare, int64, error) {
	var locations []model.SocialLocationShare
	var count int64

	// 使用Haversine公式计算距离
	// 地球半径（单位：公里）
	earthRadius := 6371.0

	// 使用参数化查询，避免SQL注入风险
	distanceExpr := gorm.Expr(
		"(? * acos(cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) + sin(radians(?)) * sin(radians(latitude))))",
		earthRadius, lat, lng, lat,
	)

	offset := (page - 1) * size

	// 构建可见性条件
	visibilityCondition := r.db.Where("visibility = 1").Or(
		"visibility = 2 AND user_id IN (SELECT target_id FROM social_relations WHERE user_id = ? AND type = 1 AND status = 1)", userID,
	)

	// 计算总数，不使用Model()方法，直接使用Table
	err := r.db.Table("social_location_shares").
		Where("(", distanceExpr, " <= ?)", radius).
		Where(visibilityCondition).
		Where("expire_time IS NULL OR expire_time > NOW()").
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	err = r.db.Table("social_location_shares").
		Select("*, ", distanceExpr, " AS distance").
		Where("(", distanceExpr, " <= ?)", radius).
		Where(visibilityCondition).
		Where("expire_time IS NULL OR expire_time > NOW()").
		Order("distance").
		Offset(offset).Limit(size).
		Find(&locations).Error
	if err != nil {
		return nil, 0, err
	}

	return locations, count, nil
}
