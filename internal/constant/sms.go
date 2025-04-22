package constant

// 短信相关常量

// SMSType 短信类型常量
type SMSType string

// 短信类型常量
const (
	// SMSTypeVerification 验证码短信
	SMSTypeVerification SMSType = "verification"
	// SMSTypeNotification 通知短信
	SMSTypeNotification SMSType = "notification"
	// SMSTypeMarketing 营销短信
	SMSTypeMarketing SMSType = "marketing"
	// SMSTypeOther 其他类型短信
	SMSTypeOther SMSType = "other"
)

// 短信状态常量
const (
	// SMSStatusSuccess 发送成功
	SMSStatusSuccess = "success"
	// SMSStatusFailed 发送失败
	SMSStatusFailed = "failed"
)

// 阿里云短信相关常量
const (
	// AliyunSMSDefaultEndpoint 阿里云短信默认接入点
	AliyunSMSDefaultEndpoint = "dysmsapi.aliyuncs.com"
)
