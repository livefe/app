package cos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"app/config"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// TencentCOSProvider 腾讯云对象存储服务提供商，实现了StorageProvider接口
type TencentCOSProvider struct {
	client *cos.Client
	config config.TencentCOSConfig
}

// NewTencentCOSProvider 创建腾讯云对象存储服务提供商实例
func NewTencentCOSProvider() (*TencentCOSProvider, error) {
	// 获取COS配置
	cosConfig := config.GetCOSConfig()

	// 创建客户端
	client, err := createTencentClient(cosConfig.Tencent)
	if err != nil {
		return nil, err
	}

	return &TencentCOSProvider{
		client: client,
		config: cosConfig.Tencent,
	}, nil
}

// createTencentClient 初始化腾讯云对象存储服务客户端
func createTencentClient(cfg config.TencentCOSConfig) (*cos.Client, error) {
	// 确保默认存储桶已设置
	if cfg.DefaultBucket == "" {
		return nil, fmt.Errorf("默认存储桶未配置")
	}

	// 将 COS 服务的 URL 解析为一个 URL 结构体
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.DefaultBucket, cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("解析腾讯云COS URL失败: %v", err)
	}

	// 基于 URL 创建 COS 客户端
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	return client, nil
}

// getBucketClient 获取指定存储桶的客户端
func (p *TencentCOSProvider) getBucketClient(bucket string) (*cos.Client, error) {
	// 如果未指定存储桶，则使用默认存储桶
	if bucket == "" {
		bucket = p.config.DefaultBucket
	}

	// 获取存储桶对象
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, p.config.Region))
	if err != nil {
		return nil, fmt.Errorf("解析存储桶URL失败: %v", err)
	}

	// 创建并返回客户端
	return cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  p.config.SecretID,
			SecretKey: p.config.SecretKey,
		},
	}), nil
}

// UploadFile 上传文件，实现StorageProvider接口
func (p *TencentCOSProvider) UploadFile(bucket, objectKey string, reader io.Reader, contentType string) (string, error) {
	// 获取存储桶客户端
	bucketClient, err := p.getBucketClient(bucket)
	if err != nil {
		return "", err
	}

	// 记录最终使用的桶名（可能是默认桶）
	if bucket == "" {
		bucket = p.config.DefaultBucket
	}

	// 上传选项
	options := &cos.ObjectPutOptions{}
	if contentType != "" {
		options.ObjectPutHeaderOptions = &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		}
	}

	// 上传文件
	_, err = bucketClient.Object.Put(context.Background(), objectKey, reader, options)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %v", err)
	}

	// 返回文件URL
	return p.getFileURL(bucket, objectKey), nil
}

// DownloadFile 下载文件，实现StorageProvider接口
func (p *TencentCOSProvider) DownloadFile(bucket, objectKey string, writer io.Writer) error {
	// 获取存储桶客户端
	bucketClient, err := p.getBucketClient(bucket)
	if err != nil {
		return err
	}

	// 下载文件
	resp, err := bucketClient.Object.Get(context.Background(), objectKey, nil)
	if err != nil {
		return fmt.Errorf("下载文件失败: %v", err)
	}
	defer resp.Body.Close()

	// 将响应内容写入writer
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件内容失败: %v", err)
	}

	return nil
}

// DeleteFile 删除文件，实现StorageProvider接口
func (p *TencentCOSProvider) DeleteFile(bucket, objectKey string) error {
	// 获取存储桶客户端
	bucketClient, err := p.getBucketClient(bucket)
	if err != nil {
		return err
	}

	// 删除文件
	_, err = bucketClient.Object.Delete(context.Background(), objectKey)
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// GetFileURL 获取文件访问URL，实现StorageProvider接口
func (p *TencentCOSProvider) GetFileURL(bucket, objectKey string, expires time.Duration) (string, error) {
	// 如果未指定存储桶，则使用默认存储桶
	if bucket == "" {
		bucket = p.config.DefaultBucket
	}

	// 如果过期时间为0，则返回永久URL（可能使用自定义域名）
	if expires == 0 {
		return p.getFileURL(bucket, objectKey), nil
	}

	// 对于需要预签名的URL，必须使用COS官方域名
	// 获取存储桶客户端
	bucketClient, err := p.getBucketClient(bucket)
	if err != nil {
		return "", err
	}

	// 生成预签名URL
	presignedURL, err := bucketClient.Object.GetPresignedURL(
		context.Background(),
		http.MethodGet,
		objectKey,
		p.config.SecretID,
		p.config.SecretKey,
		expires,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %v", err)
	}

	return presignedURL.String(), nil
}

// ListFiles 列出文件，实现StorageProvider接口
func (p *TencentCOSProvider) ListFiles(bucket, prefix string) ([]FileInfo, error) {
	// 获取存储桶客户端
	bucketClient, err := p.getBucketClient(bucket)
	if err != nil {
		return nil, err
	}

	// 记录最终使用的桶名（可能是默认桶）
	if bucket == "" {
		bucket = p.config.DefaultBucket
	}

	// 列出对象
	opt := &cos.BucketGetOptions{
		Prefix: prefix,
	}
	result, _, err := bucketClient.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil, fmt.Errorf("列出文件失败: %v", err)
	}

	// 转换为通用文件信息结构
	files := make([]FileInfo, 0, len(result.Contents))
	for _, item := range result.Contents {
		lastModified, _ := time.Parse(time.RFC3339, item.LastModified)
		files = append(files, FileInfo{
			Key:          item.Key,
			Size:         item.Size,
			LastModified: lastModified,
			ETag:         item.ETag,
			StorageClass: item.StorageClass,
		})
	}

	return files, nil
}

// getFileURL 获取文件的永久URL
func (p *TencentCOSProvider) getFileURL(bucket, objectKey string) string {
	// 检查是否启用了自定义域名映射且该桶有配置自定义域名
	if p.config.UseDomainMap && p.config.Buckets != nil {
		if customDomain, exists := p.config.Buckets[bucket]; exists && customDomain != "" {
			// 使用自定义域名
			return fmt.Sprintf("https://%s/%s", customDomain, objectKey)
		}
	}

	// 使用默认COS域名
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", bucket, p.config.Region, objectKey)
}
