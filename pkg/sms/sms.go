// Package sms 提供短信发送服务的统一接口和实现
// 支持多种短信服务提供商，当前实现了阿里云短信服务
package sms

// SMSProvider 短信服务提供商接口
// 所有短信服务提供商实现都需要实现此接口，便于扩展其他短信服务
type SMSProvider interface {
	// SendSMS 发送短信，接收通用请求参数，返回通用响应
	SendSMS(request SMSRequest) (*SMSResponse, error)
}

// SMSRequest 通用短信请求参数结构体
type SMSRequest struct {
	PhoneNumbers  string            // 接收短信的手机号码，支持对多个手机号码发送短信，手机号码之间以英文逗号（,）分隔
	SignName      string            // 短信签名名称，不同服务商可能有不同要求
	TemplateCode  string            // 短信模板ID，需要在服务商平台预先创建
	TemplateParam map[string]string // 短信模板变量对应的实际值，用于替换模板中的变量
}

// SMSResponse 通用短信发送响应结构体
// 统一不同服务商的响应格式，方便调用方处理结果
type SMSResponse struct {
	RequestId     string // 请求ID，用于问题排查
	Code          string // 状态码，表示请求处理结果
	Message       string // 状态码的描述信息
	BizId         string // 发送回执ID，可用于查询短信发送状态
	RecommendInfo string // 错误时的诊断信息，帮助解决问题
}

// SMSClient 短信客户端结构体
// 提供统一的短信发送接口，内部使用具体的服务提供商实现
type SMSClient struct {
	provider SMSProvider // 短信服务提供商实现
}

// NewSMSClient 创建短信客户端实例
// 参数:
//   - provider: 实现了SMSProvider接口的短信服务提供商
// 返回:
//   - 短信客户端指针
func NewSMSClient(provider SMSProvider) *SMSClient {
	return &SMSClient{
		provider: provider,
	}
}

// SendSMS 发送短信
// 参数:
//   - request: 短信请求参数
// 返回:
//   - 短信发送响应指针
//   - 可能的错误
// 内部委托给具体的短信服务提供商实现
func (c *SMSClient) SendSMS(request SMSRequest) (*SMSResponse, error) {
	return c.provider.SendSMS(request)
}

// GetSMSClient 获取默认的短信客户端
// 当前默认使用阿里云短信服务，未来可扩展为根据配置选择不同服务商
// 返回:
//   - 短信客户端指针
//   - 可能的错误
func GetSMSClient() (*SMSClient, error) {
	// 创建阿里云短信服务提供商
	provider, err := NewAliyunSMSProvider()
	if err != nil {
		return nil, err
	}
	// 创建并返回短信客户端
	return NewSMSClient(provider), nil
}
