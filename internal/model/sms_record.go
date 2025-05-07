package model

import (
	"time"

	"app/internal/constant"

	"gorm.io/gorm"
)

// 使用常量包中定义的SMSType类型

// SMSRecord 短信记录模型
// 用于记录所有类型的短信发送记录，包括验证码、通知和营销短信
type SMSRecord struct {
	// 基本标识信息
	ID          uint   `gorm:"primaryKey;comment:记录ID，主键" json:"id"`
	PhoneNumber string `gorm:"size:20;comment:接收短信的手机号" json:"phone_number"`

	// 短信内容信息
	Type          constant.SMSType `gorm:"size:20;comment:短信类型" json:"type"`
	Content       string           `gorm:"size:1000;comment:短信内容" json:"content"`
	TemplateCode  string           `gorm:"size:100;comment:短信模板代码" json:"template_code"`
	TemplateParam string           `gorm:"size:1000;comment:短信模板参数，JSON格式" json:"template_param"`

	// 发送状态信息
	Status       string `gorm:"size:20;comment:发送状态：success-成功，failed-失败" json:"status"`
	ErrorMessage string `gorm:"size:500;comment:错误信息" json:"error_message"`
	RequestId    string `gorm:"size:100;comment:请求ID" json:"request_id"`
	BizId        string `gorm:"size:100;comment:发送回执ID" json:"biz_id"`

	// 时间信息
	CreatedAt time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;comment:删除时间，软删除" json:"-"`
}
