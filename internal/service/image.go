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
	// UploadTempImage 上传临时图片（不关联具体模块）
	UploadTempImage(ctx context.Context, userID uint, reader io.Reader, filename string, size int64) (*model.TempImage, error)
	// UploadMultipleTempImages 批量上传临时图片
	UploadMultipleTempImages(ctx context.Context, userID uint, files []io.Reader, filenames []string, sizes []int64) ([]model.TempImage, []error)
	// UploadAvatar 上传用户头像
	UploadAvatar(ctx context.Context, userID uint, reader io.Reader, filename string) (string, error)
	// MoveImageToPost 将临时图片移动到动态并关联
	MoveImageToPost(ctx context.Context, imageID, postID, userID uint) (*model.PostImage, error)
}

// imageService 图片服务实现
type imageService struct {
	postImageRepo repository.PostImageRepository
	tempImageRepo repository.TempImageRepository
	userRepo      repository.UserRepository
	cosClient     *cos.StorageClient
	postRepo      repository.PostRepository
}

// NewImageService 创建图片服务实例
func NewImageService(
	postImageRepo repository.PostImageRepository,
	tempImageRepo repository.TempImageRepository,
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
) (ImageService, error) {
	// 获取COS客户端
	cosClient, err := cos.GetStorageClient()
	if err != nil {
		return nil, fmt.Errorf("获取COS客户端失败: %w", err)
	}

	return &imageService{
		postImageRepo: postImageRepo,
		tempImageRepo: tempImageRepo,
		userRepo:      userRepo,
		postRepo:      postRepo,
		cosClient:     cosClient,
	}, nil
}

// UploadTempImage 上传临时图片（通用接口，不关联具体模块）
func (s *imageService) UploadTempImage(ctx context.Context, userID uint, reader io.Reader, filename string, size int64) (*model.TempImage, error) {
	// 生成临时图片的对象键名
	objectKey := generateTempImageObjectKey(userID, filename)

	// 获取文件内容类型
	contentType := getContentTypeByFilename(filename)

	// 上传到COS
	url, err := s.cosClient.UploadFile("", objectKey, reader, contentType)
	if err != nil {
		return nil, fmt.Errorf("上传临时图片到COS失败: %w", err)
	}

	// 创建临时图片记录
	tempImage := &model.TempImage{
		UserID:      userID,
		ObjectKey:   objectKey,
		URL:         url,
		Bucket:      "", // 使用默认存储桶
		Size:        size,
		ContentType: contentType,
	}

	// 保存到数据库
	err = s.tempImageRepo.CreateTempImage(tempImage)
	if err != nil {
		return nil, fmt.Errorf("保存临时图片记录失败: %w", err)
	}

	return tempImage, nil
}

// UploadMultipleTempImages 批量上传临时图片
func (s *imageService) UploadMultipleTempImages(ctx context.Context, userID uint, files []io.Reader, filenames []string, sizes []int64) ([]model.TempImage, []error) {
	// 检查参数长度是否一致
	if len(files) != len(filenames) || len(files) != len(sizes) {
		return nil, []error{fmt.Errorf("参数数量不匹配")}
	}

	// 存储上传结果
	results := make([]model.TempImage, 0, len(files))
	errs := make([]error, len(files))

	// 循环上传每个文件
	for i, file := range files {
		tempImage, err := s.UploadTempImage(ctx, userID, file, filenames[i], sizes[i])
		if err != nil {
			// 记录错误
			errs[i] = fmt.Errorf("上传图片 %s 失败: %w", filenames[i], err)
		} else {
			// 添加成功结果
			results = append(results, *tempImage)
			errs[i] = nil
		}
	}

	return results, errs
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

// MoveImageToPost 将临时图片移动到动态并关联
func (s *imageService) MoveImageToPost(ctx context.Context, imageID, postID, userID uint) (*model.PostImage, error) {
	// 查找临时图片
	tempImage, err := s.tempImageRepo.FindByID(imageID)
	if err != nil {
		return nil, fmt.Errorf("查找临时图片记录失败: %w", err)
	}

	// 验证图片所有权
	if tempImage.UserID != userID {
		return nil, fmt.Errorf("无权操作此图片")
	}

	// 验证动态是否存在
	_, err = s.postRepo.GetPost(postID)
	if err != nil {
		return nil, fmt.Errorf("动态不存在: %w", err)
	}

	// 从临时对象键名中提取文件名
	oldObjectKey := tempImage.ObjectKey
	filename := filepath.Base(oldObjectKey)

	// 生成新的对象键名（动态图片的最终位置）
	newObjectKey := generatePostImageObjectKey(userID, postID, filename)

	// 在COS中复制文件到新位置
	err = s.cosClient.CopyFile("", oldObjectKey, "", newObjectKey)
	if err != nil {
		return nil, fmt.Errorf("移动图片到最终位置失败: %w", err)
	}

	// 获取新文件的URL
	newURL, err := s.cosClient.GetFileURL("", newObjectKey, 0) // 使用永久URL
	if err != nil {
		// 如果获取URL失败，使用替代方法
		newURL = strings.Replace(tempImage.URL, oldObjectKey, newObjectKey, 1)
	}

	// 创建动态图片记录
	postImage := &model.PostImage{
		PostID:      postID,
		UserID:      userID,
		ObjectKey:   newObjectKey,
		URL:         newURL,
		Bucket:      tempImage.Bucket,
		Size:        tempImage.Size,
		ContentType: tempImage.ContentType,
	}

	// 保存到数据库
	err = s.postImageRepo.CreatePostImage(postImage)
	if err != nil {
		return nil, fmt.Errorf("创建动态图片记录失败: %w", err)
	}

	// 删除临时图片记录
	err = s.tempImageRepo.DeleteTempImage(imageID)
	if err != nil {
		// 仅记录错误，不影响主流程
		fmt.Printf("删除临时图片记录失败: %v\n", err)
	}

	return postImage, nil
}

// 生成动态图片的对象键名
func generatePostImageObjectKey(userID, postID uint, filename string) string {
	extension := filepath.Ext(filename)
	timestamp := time.Now().UnixNano() / 1e6 // 毫秒级时间戳
	return fmt.Sprintf("posts/%d/%d/%d%s", userID, postID, timestamp, extension)
}

// 生成临时图片的对象键名
func generateTempImageObjectKey(userID uint, filename string) string {
	extension := filepath.Ext(filename)
	timestamp := time.Now().UnixNano() / 1e6 // 毫秒时间戳
	return fmt.Sprintf("temp/%d/%d%s", userID, timestamp, extension)
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
