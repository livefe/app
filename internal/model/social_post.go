package model

import (
	"time"

	"gorm.io/gorm"
)

// SocialPost 社交动态模型
// 存储用户发布的社交动态内容
type SocialPost struct {
	ID         uint           `gorm:"primaryKey;comment:动态ID，主键" json:"id"`
	UserID     uint           `gorm:"index:idx_user_post;comment:用户ID" json:"user_id"`
	Content    string         `gorm:"size:2000;comment:动态内容" json:"content"`
	Images     string         `gorm:"size:1000;comment:图片URL，多个以逗号分隔" json:"images"`
	LocationID *uint          `gorm:"index;comment:关联的社交位置分享ID" json:"location_id"`
	Visibility int            `gorm:"type:smallint;default:1;index:idx_visibility_created;comment:可见性：1-公开，2-仅好友，3-私密" json:"visibility"`
	Likes      int            `gorm:"default:0;comment:点赞数" json:"likes"`
	Comments   int            `gorm:"default:0;comment:评论数" json:"comments"`
	CreatedAt  time.Time      `gorm:"type:datetime;index:idx_visibility_created;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"type:datetime;comment:删除时间，软删除" json:"-"`
}
