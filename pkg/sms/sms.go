// Package sms 提供短信发送服务的统一接口和实现，支持多种短信服务提供商
package sms

import "fmt"

// SMSProvider 短信服务提供商接口，所有短信服务提供商都需要实现此接口
type SMSProvider interface {
	// SendSMS 发送短信，接收通用请求参数，返回通用响应
	SendSMS(request SMSRequest) (*SMSResponse, error)
}

// SMSRequest 通用短信请求参数结构体，统一不同服务商的请求格式
type SMSRequest struct {
	PhoneNumbers  string            // 接收短信的手机号码，多个号码以英文逗号分隔
	SignName      string            // 短信签名名称
	TemplateCode  string            // 短信模板ID
	TemplateParam map[string]string // 短信模板变量对应的实际值
}

// SMSResponse 通用短信发送响应结构体，统一不同服务商的响应格式
type SMSResponse struct {
	RequestId     string // 请求ID，用于问题排查
	Code          string // 状态码
	Message       string // 状态码的描述信息
	BizId         string // 发送回执ID
	RecommendInfo string // 错误时的诊断信息
}

// SMSClient 短信客户端结构体，提供统一的短信发送接口
type SMSClient struct {
	provider SMSProvider // 短信服务提供商实现
}

// NewSMSClient 创建短信客户端实例
// 参数: provider - 实现了SMSProvider接口的短信服务提供商
// 返回: 短信客户端指针
func NewSMSClient(provider SMSProvider) *SMSClient {
	return &SMSClient{
		provider: provider,
	}
}

// SendSMS 发送短信，内部委托给具体的短信服务提供商实现
// 参数: request - 短信请求参数
// 返回: 短信发送响应指针和可能的错误
func (c *SMSClient) SendSMS(request SMSRequest) (*SMSResponse, error) {
	return c.provider.SendSMS(request)
}

// ProviderType 短信服务提供商类型，用于标识不同的短信服务提供商
type ProviderType string

// 支持的短信服务提供商类型
const (
	AliyunProvider ProviderType = "aliyun" // 阿里云短信服务
	// 未来可以添加更多服务商
	// TencentProvider ProviderType = "tencent" // 腾讯云短信服务
	// AWSProvider     ProviderType = "aws"     // AWS SNS短信服务
)

// GetSMSClient 获取短信客户端，根据提供的服务商类型返回对应的实例
// 参数: providerType - 短信服务提供商类型，默认为阿里云
// 返回: 短信客户端指针和可能的错误
func GetSMSClient(providerType ...ProviderType) (*SMSClient, error) {
	// 默认使用阿里云短信服务
	pType := AliyunProvider
	if len(providerType) > 0 && providerType[0] != "" {
		pType = providerType[0]
	}

	// 根据提供商类型创建对应的服务提供商
	var provider SMSProvider
	var err error

	switch pType {
	case AliyunProvider:
		provider, err = NewAliyunSMSProvider()
	// 未来可以添加更多服务商的支持
	// case TencentProvider:
	// 	provider, err = NewTencentSMSProvider()
	// case AWSProvider:
	// 	provider, err = NewAWSSMSProvider()
	default:
		return nil, fmt.Errorf("不支持的短信服务提供商类型: %s", pType)
	}

	if err != nil {
		return nil, err
	}

	// 创建并返回短信客户端
	return NewSMSClient(provider), nil
}
