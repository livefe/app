package sms

import (
	"encoding/json"
	"fmt"
	"strings"

	"app/config"
	"app/internal/constant"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// AliyunSMSProvider 阿里云短信服务提供商，实现了SMSProvider接口
type AliyunSMSProvider struct {
	client *dysmsapi20170525.Client
	config config.AliyunSMSConfig
}

// NewAliyunSMSProvider 创建阿里云短信服务提供商实例
func NewAliyunSMSProvider() (*AliyunSMSProvider, error) {
	// 获取短信配置
	smsConfig := config.GetSMSConfig()

	// 创建客户端
	client, err := createClient(smsConfig.Aliyun)
	if err != nil {
		return nil, err
	}

	return &AliyunSMSProvider{
		client: client,
		config: smsConfig.Aliyun,
	}, nil
}

// createClient 初始化阿里云短信服务客户端
func createClient(cfg config.AliyunSMSConfig) (*dysmsapi20170525.Client, error) {
	// 工程代码泄露可能会导致 AccessKey 泄露，并威胁账号下所有资源的安全性
	// 建议使用更安全的 STS 方式，更多鉴权访问方式请参见：https://help.aliyun.com/document_detail/378661.html
	clientConfig := &openapi.Config{
		AccessKeyId:     tea.String(cfg.AccessKeyID),
		AccessKeySecret: tea.String(cfg.AccessKeySecret),
	}

	// 设置API接入地址
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = constant.AliyunSMSDefaultEndpoint // 使用常量定义的默认值
	}
	clientConfig.Endpoint = tea.String(endpoint)

	return dysmsapi20170525.NewClient(clientConfig)
}

// SendSMS 发送短信，实现SMSProvider接口
func (c *AliyunSMSProvider) SendSMS(req SMSRequest) (*SMSResponse, error) {
	// 构建请求
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers: tea.String(req.PhoneNumbers),
		TemplateCode: tea.String(req.TemplateCode),
	}

	// 使用配置中的签名名称（如果请求中未指定）
	signName := req.SignName
	if signName == "" {
		signName = c.config.SignName
	}
	sendSmsRequest.SignName = tea.String(signName)

	// 处理模板参数
	if req.TemplateParam != nil && len(req.TemplateParam) > 0 {
		templateParamJSON, err := json.Marshal(req.TemplateParam)
		if err != nil {
			return nil, fmt.Errorf("模板参数序列化失败: %v", err)
		}
		sendSmsRequest.TemplateParam = tea.String(string(templateParamJSON))
	}

	// 设置运行时选项
	runtime := &util.RuntimeOptions{}

	// 发送短信
	response, err := c.client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		return handleError(err)
	}

	// 构建响应
	return &SMSResponse{
		RequestId: tea.StringValue(response.Body.RequestId),
		Code:      tea.StringValue(response.Body.Code),
		Message:   tea.StringValue(response.Body.Message),
		BizId:     tea.StringValue(response.Body.BizId),
	}, nil
}

// handleError 处理错误
func handleError(err error) (*SMSResponse, error) {
	var sdkError *tea.SDKError
	if _t, ok := err.(*tea.SDKError); ok {
		sdkError = _t
	} else {
		sdkError = &tea.SDKError{Message: tea.String(err.Error())}
	}

	// 构建错误响应
	response := &SMSResponse{
		Message: tea.StringValue(sdkError.Message),
	}

	// 尝试解析诊断信息
	if sdkError.Data != nil {
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(sdkError.Data)))
		d.DisallowUnknownFields()
		if decodeErr := d.Decode(&data); decodeErr == nil {
			if m, ok := data.(map[string]interface{}); ok {
				if recommend, ok := m["Recommend"]; ok {
					response.RecommendInfo = fmt.Sprintf("%v", recommend)
				}
			}
		}
	}

	return response, err
}
