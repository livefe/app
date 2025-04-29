package model

import (
	"time"

	"gorm.io/gorm"
)

// SocialComment 社交评论模型
// 存储用户对动态的评论
type SocialComment struct {
	ID        uint           `gorm:"primaryKey;comment:评论ID，主键" json:"id"`
	PostID    uint           `gorm:"index:idx_post_comment,priority:1;comment:动态ID" json:"post_id"`
	UserID    uint           `gorm:"index;comment:评论用户ID" json:"user_id"`
	Content   string         `gorm:"size:500;comment:评论内容" json:"content"`
	ParentID  *uint          `gorm:"index;comment:父评论ID，用于回复功能" json:"parent_id"`
	CreatedAt time.Time      `gorm:"type:datetime;index:idx_post_comment,priority:2;index;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;comment:删除时间，软删除" json:"-"`
}
