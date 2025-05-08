package model

import (
	"time"

	"gorm.io/gorm"
)

// UserFriend 好友关系模型
// 存储用户之间的好友关系
// 采用双记录模式，每个好友关系在数据库中存储为两条记录
type UserFriend struct {
	// 基本标识信息
	ID uint `gorm:"primaryKey;comment:好友关系ID，主键" json:"id"`

	// 关系信息
	UserID   uint `gorm:"comment:用户ID，记录所有者" json:"user_id"`
	TargetID uint `gorm:"comment:目标用户ID，好友对象" json:"target_id"`

	// 状态信息
	Status int `gorm:"type:smallint;default:0;comment:好友状态：0-待确认，1-已确认" json:"status"`
	// 关系方向，表示是发起方还是接收方
	Direction int `gorm:"type:smallint;default:0;comment:关系方向：0-发起方，1-接收方" json:"direction"`

	// 时间信息
	CreatedAt time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;comment:删除时间，软删除" json:"-"`
}