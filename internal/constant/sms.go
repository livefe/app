package constant

// SMSType 短信类型
type SMSType string

const (
	// 验证码短信
	SMSTypeVerification SMSType = "verification"
	// 通知短信
	SMSTypeNotification SMSType = "notification"
	// 营销短信
	SMSTypeMarketing SMSType = "marketing"
	// 其他类型短信
	SMSTypeOther SMSType = "other"
)

// 短信状态常量
const (
	// 发送成功
	SMSStatusSuccess = "success"
	// 发送失败
	SMSStatusFailed = "failed"
)

// 阿里云短信相关常量
const (
	// 阿里云短信默认接入点
	AliyunSMSDefaultEndpoint = "dysmsapi.aliyuncs.com"
)
