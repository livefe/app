package model

import (
	"time"

	"gorm.io/gorm"
)

// Post 动态模型
// 存储用户发布的动态内容
type Post struct {
	ID         uint           `gorm:"primaryKey;comment:动态ID，主键" json:"id"`
	UserID     uint           `gorm:"comment:用户ID" json:"user_id"`
	Content    string         `gorm:"size:2000;comment:动态内容" json:"content"`
	Visibility int            `gorm:"type:smallint;default:1;comment:可见性：1-公开，2-仅好友，3-私密" json:"visibility"`
	PostImages []PostImage    `gorm:"foreignKey:PostID" json:"-"` // 关联的图片列表
	Likes      int            `gorm:"default:0;comment:点赞数" json:"likes"`
	Comments   int            `gorm:"default:0;comment:评论数" json:"comments"`
	CreatedAt  time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"type:datetime;comment:删除时间" json:"-"`
}
