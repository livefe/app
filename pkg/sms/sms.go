package sms

// SMSProvider 短信服务提供商接口
// 所有短信服务提供商都需要实现这个接口
type SMSProvider interface {
	// SendSMS 发送短信
	// 参数为通用的短信请求，返回通用的短信响应
	SendSMS(request SMSRequest) (*SMSResponse, error)
}

// SMSRequest 通用短信请求参数
type SMSRequest struct {
	PhoneNumbers  string            // 接收短信的手机号码，支持对多个手机号码发送短信，手机号码之间以英文逗号（,）分隔
	SignName      string            // 短信签名名称
	TemplateCode  string            // 短信模板ID
	TemplateParam map[string]string // 短信模板变量对应的实际值
}

// SMSResponse 通用短信发送响应
type SMSResponse struct {
	RequestId     string // 请求ID
	Code          string // 状态码
	Message       string // 状态码的描述
	BizId         string // 发送回执ID
	RecommendInfo string // 错误时的诊断信息
}

// SMSClient 短信客户端
// 用于获取短信服务提供商实例并发送短信
type SMSClient struct {
	provider SMSProvider
}

// NewSMSClient 创建短信客户端
// 参数为短信服务提供商实例
func NewSMSClient(provider SMSProvider) *SMSClient {
	return &SMSClient{
		provider: provider,
	}
}

// SendSMS 发送短信
// 委托给具体的短信服务提供商实现
func (c *SMSClient) SendSMS(request SMSRequest) (*SMSResponse, error) {
	return c.provider.SendSMS(request)
}

// GetDefaultSMSClient 获取默认的短信客户端
// 目前默认使用阿里云短信服务，未来可以根据配置动态选择不同的短信服务提供商
func GetDefaultSMSClient() (*SMSClient, error) {
	// 创建阿里云短信服务提供商实例
	provider, err := NewAliyunSMSProvider()
	if err != nil {
		return nil, err
	}

	// 创建短信客户端
	return NewSMSClient(provider), nil
}
