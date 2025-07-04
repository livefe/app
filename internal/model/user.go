package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
// 存储系统用户的基本信息，包含用户的基础资料和账号状态
type User struct {
	ID        uint           `gorm:"primaryKey;comment:用户ID，主键" json:"id"`
	Username  string         `gorm:"size:50;comment:用户名，登录账号" json:"username"`
	Password  string         `gorm:"size:100;comment:密码，加密存储" json:"-"`
	Mobile    string         `gorm:"size:20;comment:手机号，用于验证码登录" json:"mobile"`
	Nickname  string         `gorm:"size:50;comment:用户昵称，显示名称" json:"nickname"`
	Avatar    string         `gorm:"size:255;comment:用户头像URL" json:"avatar"`
	Status    int            `gorm:"type:smallint;default:1;comment:用户状态：1-正常，0-禁用" json:"status"`
	CreatedAt time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;comment:删除时间" json:"-"`
}
