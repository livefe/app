package model

import (
	"time"

	"gorm.io/gorm"
)

// Follower 粉丝关注模型
// 存储用户之间的关注关系
type Follower struct {
	// 基本标识信息
	ID uint `gorm:"primaryKey;comment:关注ID，主键" json:"id"`

	// 关系信息
	UserID   uint `gorm:"index:idx_user_follower,priority:1;index:idx_user_target_follower,priority:1,uniqueIndex;comment:用户ID，关注发起者" json:"user_id"`
	TargetID uint `gorm:"index:idx_user_follower,priority:2;index:idx_user_target_follower,priority:2,uniqueIndex;comment:目标用户ID，被关注者" json:"target_id"`

	// 时间信息
	CreatedAt time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;comment:删除时间，软删除" json:"-"`
}
