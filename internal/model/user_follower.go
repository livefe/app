package model

import (
	"time"

	"gorm.io/gorm"
)

// UserFollower 粉丝关注模型
// 存储用户之间的关注关系
type UserFollower struct {
	ID uint `gorm:"primaryKey;comment:关注ID，主键" json:"id"`

	UserID   uint `gorm:"comment:用户ID，关注发起者" json:"user_id"`
	TargetID uint `gorm:"comment:目标用户ID，被关注者" json:"target_id"`

	CreatedAt time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;comment:删除时间" json:"-"`
}
