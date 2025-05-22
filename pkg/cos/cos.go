// Package cos 提供对象存储服务的统一接口和实现，支持多种对象存储服务提供商
package cos

import (
	"fmt"
	"io"
	"time"
)

// StorageProvider 对象存储服务提供商接口，所有对象存储服务提供商都需要实现此接口
type StorageProvider interface {
	// UploadFile 上传文件
	// 参数: bucket - 存储桶名称, objectKey - 对象键, reader - 文件内容读取器, contentType - 内容类型
	// 返回: 访问URL和可能的错误
	UploadFile(bucket, objectKey string, reader io.Reader, contentType string) (string, error)

	// DownloadFile 下载文件
	// 参数: bucket - 存储桶名称, objectKey - 对象键, writer - 文件内容写入器
	// 返回: 可能的错误
	DownloadFile(bucket, objectKey string, writer io.Writer) error

	// DeleteFile 删除文件
	// 参数: bucket - 存储桶名称, objectKey - 对象键
	// 返回: 可能的错误
	DeleteFile(bucket, objectKey string) error

	// GetFileURL 获取文件访问URL
	// 参数: bucket - 存储桶名称, objectKey - 对象键, expires - URL过期时间
	// 返回: 访问URL和可能的错误
	GetFileURL(bucket, objectKey string, expires time.Duration) (string, error)

	// ListFiles 列出文件
	// 参数: bucket - 存储桶名称, prefix - 前缀
	// 返回: 文件列表和可能的错误
	ListFiles(bucket, prefix string) ([]FileInfo, error)
}

// FileInfo 文件信息结构体
type FileInfo struct {
	Key          string    // 对象键
	Size         int64     // 文件大小（字节）
	LastModified time.Time // 最后修改时间
	ETag         string    // 文件的 ETag
	StorageClass string    // 存储类型
}

// StorageRequest 通用存储请求参数结构体
type StorageRequest struct {
	Bucket      string    // 存储桶名称
	ObjectKey   string    // 对象键
	ContentType string    // 内容类型
	Expires     time.Time // 过期时间
}

// StorageClient 对象存储客户端结构体，提供统一的对象存储接口
type StorageClient struct {
	provider StorageProvider // 对象存储服务提供商实现
}

// NewStorageClient 创建对象存储客户端实例
// 参数: provider - 实现了StorageProvider接口的对象存储服务提供商
// 返回: 对象存储客户端指针
func NewStorageClient(provider StorageProvider) *StorageClient {
	return &StorageClient{
		provider: provider,
	}
}

// UploadFile 上传文件，内部委托给具体的对象存储服务提供商实现
func (c *StorageClient) UploadFile(bucket, objectKey string, reader io.Reader, contentType string) (string, error) {
	return c.provider.UploadFile(bucket, objectKey, reader, contentType)
}

// DownloadFile 下载文件，内部委托给具体的对象存储服务提供商实现
func (c *StorageClient) DownloadFile(bucket, objectKey string, writer io.Writer) error {
	return c.provider.DownloadFile(bucket, objectKey, writer)
}

// DeleteFile 删除文件，内部委托给具体的对象存储服务提供商实现
func (c *StorageClient) DeleteFile(bucket, objectKey string) error {
	return c.provider.DeleteFile(bucket, objectKey)
}

// GetFileURL 获取文件访问URL，内部委托给具体的对象存储服务提供商实现
func (c *StorageClient) GetFileURL(bucket, objectKey string, expires time.Duration) (string, error) {
	return c.provider.GetFileURL(bucket, objectKey, expires)
}

// ListFiles 列出文件，内部委托给具体的对象存储服务提供商实现
func (c *StorageClient) ListFiles(bucket, prefix string) ([]FileInfo, error) {
	return c.provider.ListFiles(bucket, prefix)
}

// ProviderType 对象存储服务提供商类型，用于标识不同的对象存储服务提供商
type ProviderType string

// 支持的对象存储服务提供商类型
const (
	TencentProvider ProviderType = "tencent" // 腾讯云对象存储
	// 未来可以添加更多服务商
	// AliyunProvider  ProviderType = "aliyun"  // 阿里云对象存储
	// AWSProvider     ProviderType = "aws"     // AWS S3对象存储
)

// GetStorageClient 获取对象存储客户端，根据提供的服务商类型返回对应的实例
// 参数: providerType - 对象存储服务提供商类型，默认为腾讯云
// 返回: 对象存储客户端指针和可能的错误
func GetStorageClient(providerType ...ProviderType) (*StorageClient, error) {
	// 默认使用腾讯云对象存储服务
	pType := TencentProvider
	if len(providerType) > 0 && providerType[0] != "" {
		pType = providerType[0]
	}

	// 根据提供商类型创建对应的服务提供商
	var provider StorageProvider
	var err error

	switch pType {
	case TencentProvider:
		provider, err = NewTencentCOSProvider()
	// 未来可以添加更多服务商的支持
	// case AliyunProvider:
	// 	provider, err = NewAliyunOSSProvider()
	// case AWSProvider:
	// 	provider, err = NewAWSS3Provider()
	default:
		return nil, fmt.Errorf("不支持的对象存储服务提供商类型: %s", pType)
	}

	if err != nil {
		return nil, err
	}

	// 创建并返回对象存储客户端
	return NewStorageClient(provider), nil
}
