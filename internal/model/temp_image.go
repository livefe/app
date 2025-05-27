package model

import (
	"time"

	"gorm.io/gorm"
)

// TempImage 临时图片模型
// 存储用户上传的临时图片信息
type TempImage struct {
	ID          uint           `gorm:"primaryKey;comment:图片ID，主键" json:"id"`
	UserID      uint           `gorm:"index;comment:用户ID" json:"user_id"`
	ObjectKey   string         `gorm:"size:255;comment:对象存储中的键名" json:"object_key"`
	URL         string         `gorm:"size:500;comment:图片访问URL" json:"url"`
	Bucket      string         `gorm:"size:100;comment:存储桶名称" json:"bucket"`
	Size        int64          `gorm:"comment:图片大小(字节)" json:"size"`
	ContentType string         `gorm:"size:50;comment:内容类型" json:"content_type"`
	CreatedAt   time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"type:datetime;comment:删除时间" json:"-"`
}
