package handler

import (
	"app/internal/service"
	"app/pkg/response"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// ImageHandler 图片处理器
type ImageHandler struct {
	imageService service.ImageService
	postService  service.PostService
}

// NewImageHandler 创建图片处理器实例
func NewImageHandler(imageService service.ImageService, postService service.PostService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
		postService:  postService,
	}
}

// UploadPostImage 上传动态图片（二进制文件方式）- 已废弃，请使用UploadTempImage和MoveImageToPost代替
func (h *ImageHandler) UploadPostImage(c *gin.Context) {
	// 返回错误提示，建议使用新的上传方式
	response.BadRequest(c, "此接口已废弃，请先使用临时图片上传接口，再移动图片到动态", nil)
}

// UploadTempImage 上传临时图片（通用接口，不关联具体模块）
func (h *ImageHandler) UploadTempImage(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		response.BadRequest(c, "获取上传文件失败", err)
		return
	}

	// 检查文件大小（可选，例如限制为10MB）
	if file.Size > 10*1024*1024 {
		response.BadRequest(c, "文件大小超过限制", nil)
		return
	}

	// 检查文件类型（可选）
	ext := filepath.Ext(file.Filename)
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !validExts[ext] {
		response.BadRequest(c, "不支持的文件类型", nil)
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "打开上传文件失败", err)
		return
	}
	defer src.Close()

	// 上传临时图片
	tempImage, err := h.imageService.UploadTempImage(c.Request.Context(), userID.(uint), src, file.Filename, file.Size)
	if err != nil {
		response.InternalServerError(c, "上传图片失败", err)
		return
	}

	response.Success(c, "上传图片成功", gin.H{
		"id":           tempImage.ID,
		"url":          tempImage.URL,
		"size":         tempImage.Size,
		"content_type": tempImage.ContentType,
		"filename":     filepath.Base(file.Filename),
	})
}

// MoveImageToPost 已废弃，请在创建动态时直接关联图片
func (h *ImageHandler) MoveImageToPost(c *gin.Context) {
	// 返回错误提示，建议使用新的方式
	response.BadRequest(c, "此接口已废弃，请在创建动态时直接关联图片ID", nil)
}

// UploadAvatar 上传用户头像（二进制文件方式）
func (h *ImageHandler) UploadAvatar(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "获取上传文件失败", err)
		return
	}

	// 检查文件大小（可选，例如限制为5MB）
	if file.Size > 5*1024*1024 {
		response.BadRequest(c, "文件大小超过限制", nil)
		return
	}

	// 检查文件类型（可选）
	ext := filepath.Ext(file.Filename)
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !validExts[ext] {
		response.BadRequest(c, "不支持的文件类型", nil)
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "打开上传文件失败", err)
		return
	}
	defer src.Close()

	// 上传头像
	avatarURL, err := h.imageService.UploadAvatar(c.Request.Context(), userID.(uint), src, file.Filename)
	if err != nil {
		response.InternalServerError(c, "上传头像失败", err)
		return
	}

	response.Success(c, "上传头像成功", gin.H{
		"avatar_url": avatarURL,
	})
}
