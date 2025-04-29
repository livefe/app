package model

import (
	"time"

	"gorm.io/gorm"
)

// SocialLocationShare 位置分享模型
// 存储用户分享的位置信息
type SocialLocationShare struct {
	ID          uint           `gorm:"primaryKey;comment:位置分享ID，主键" json:"id"`
	UserID      uint           `gorm:"comment:用户ID" json:"user_id"`
	Latitude    float64        `gorm:"type:decimal(10,7);comment:纬度" json:"latitude"`
	Longitude   float64        `gorm:"type:decimal(10,7);comment:经度" json:"longitude"`
	Address     string         `gorm:"size:255;comment:地址描述" json:"address"`
	Description string         `gorm:"size:500;comment:位置描述" json:"description"`
	Visibility  int            `gorm:"type:smallint;default:1;comment:可见性：1-公开，2-仅好友，3-私密" json:"visibility"`
	ExpireTime  *time.Time     `gorm:"type:datetime;comment:过期时间，为空表示永久有效" json:"expire_time"`
	CreatedAt   time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"type:datetime;comment:删除时间，软删除" json:"-"`
	Distance    float64        `gorm:"-" json:"distance"` // 距离字段，不映射到数据库
}
