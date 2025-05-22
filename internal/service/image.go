package service

import (
	"app/internal/model"
	"app/internal/repository"
	"app/pkg/cos"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// ImageService 图片服务接口
type ImageService interface {
	// UploadPostImage 上传动态图片
	UploadPostImage(ctx context.Context, postID, userID uint, reader io.Reader, filename string, size int64) (*model.PostImage, error)
	// UploadAvatar 上传用户头像
	UploadAvatar(ctx context.Context, userID uint, reader io.Reader, filename string) (string, error)
	// GetPostImages 获取动态图片
	GetPostImages(ctx context.Context, postID uint) ([]model.PostImage, error)
}

// imageService 图片服务实现
type imageService struct {
	postImageRepo repository.PostImageRepository
	userRepo      repository.UserRepository
	cosClient     *cos.StorageClient
}

// NewImageService 创建图片服务实例
func NewImageService(
	postImageRepo repository.PostImageRepository,
	userRepo repository.UserRepository,
) (ImageService, error) {
	// 获取COS客户端
	cosClient, err := cos.GetStorageClient()
	if err != nil {
		return nil, fmt.Errorf("获取COS客户端失败: %w", err)
	}

	return &imageService{
		postImageRepo: postImageRepo,
		userRepo:      userRepo,
		cosClient:     cosClient,
	}, nil
}

// UploadPostImage 上传动态图片
func (s *imageService) UploadPostImage(ctx context.Context, postID, userID uint, reader io.Reader, filename string, size int64) (*model.PostImage, error) {
	// 生成对象键名
	objectKey := generatePostImageObjectKey(userID, postID, filename)

	// 获取文件内容类型
	contentType := getContentTypeByFilename(filename)

	// 上传到COS
	url, err := s.cosClient.UploadFile("", objectKey, reader, contentType)
	if err != nil {
		return nil, fmt.Errorf("上传图片到COS失败: %w", err)
	}

	// 创建图片记录
	postImage := &model.PostImage{
		PostID:      postID,
		UserID:      userID,
		ObjectKey:   objectKey,
		URL:         url,
		Bucket:      "", // 使用默认存储桶
		Size:        size,
		ContentType: contentType,
		// 宽高信息可以在后续处理中添加
	}

	// 保存到数据库
	err = s.postImageRepo.CreatePostImage(postImage)
	if err != nil {
		return nil, fmt.Errorf("保存图片记录失败: %w", err)
	}

	return postImage, nil
}

// UploadAvatar 上传用户头像
func (s *imageService) UploadAvatar(ctx context.Context, userID uint, reader io.Reader, filename string) (string, error) {
	// 查找用户
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", fmt.Errorf("查找用户失败: %w", err)
	}

	// 生成对象键名
	objectKey := generateAvatarObjectKey(userID, filename)

	// 获取文件内容类型
	contentType := getContentTypeByFilename(filename)

	// 上传到COS（使用默认存储桶）
	url, err := s.cosClient.UploadFile("", objectKey, reader, contentType)
	if err != nil {
		return "", fmt.Errorf("上传头像到COS失败: %w", err)
	}

	// 更新用户头像信息
	user.Avatar = url
	user.AvatarObjectKey = objectKey
	user.AvatarBucket = ""

	err = s.userRepo.Update(user)
	if err != nil {
		return "", fmt.Errorf("更新用户头像失败: %w", err)
	}

	return url, nil
}

// GetPostImages 获取动态图片
func (s *imageService) GetPostImages(ctx context.Context, postID uint) ([]model.PostImage, error) {
	return s.postImageRepo.GetPostImages(postID)
}

// 生成动态图片的对象键名
func generatePostImageObjectKey(userID, postID uint, filename string) string {
	extension := filepath.Ext(filename)
	timestamp := time.Now().UnixNano() / 1e6 // 毫秒级时间戳
	return fmt.Sprintf("posts/%d/%d/%d%s", userID, postID, timestamp, extension)
}

// 生成用户头像的对象键名
func generateAvatarObjectKey(userID uint, filename string) string {
	extension := filepath.Ext(filename)
	timestamp := time.Now().UnixNano() / 1e6 // 毫秒级时间戳
	return fmt.Sprintf("avatars/%d/%d%s", userID, timestamp, extension)
}

// 根据文件名获取内容类型
func getContentTypeByFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
