package model

import (
	"time"

	"gorm.io/gorm"
)

// PostComment 动态评论模型
// 存储用户对动态的评论
type PostComment struct {
	// 基本标识信息
	ID uint `gorm:"primaryKey;comment:评论ID，主键" json:"id"`

	// 关联信息
	PostID   uint  `gorm:"comment:动态ID" json:"post_id"`
	UserID   uint  `gorm:"comment:评论用户ID" json:"user_id"`
	ParentID *uint `gorm:"comment:父评论ID，用于回复功能" json:"parent_id"`

	// 内容信息
	Content string `gorm:"size:500;comment:评论内容" json:"content"`

	// 时间信息
	CreatedAt time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;comment:删除时间，软删除" json:"-"`
}
