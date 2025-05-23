package handler

import (
	"app/internal/service"
	"app/pkg/response"
	"fmt"
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

// UploadPostImage 上传动态图片（二进制文件方式）
func (h *ImageHandler) UploadPostImage(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 获取动态ID
	postIDStr := c.PostForm("post_id")
	if postIDStr == "" {
		response.BadRequest(c, "缺少动态ID参数", nil)
		return
	}

	// 将postID转换为uint
	var postID uint
	if _, err := fmt.Sscanf(postIDStr, "%d", &postID); err != nil {
		response.BadRequest(c, "动态ID格式错误", err)
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

	// 直接上传图片到最终位置，不再使用临时目录
	postImage, err := h.imageService.UploadPostImage(c.Request.Context(), postID, userID.(uint), src, file.Filename, file.Size)
	if err != nil {
		response.InternalServerError(c, "上传图片失败", err)
		return
	}

	response.Success(c, "上传图片成功", gin.H{
		"id":  postImage.ID,
		"url": postImage.URL,
	})
}

// UploadTempImage 上传临时图片（不关联动态ID）
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
	postImage, err := h.imageService.UploadTempImage(c.Request.Context(), userID.(uint), src, file.Filename, file.Size)
	if err != nil {
		response.InternalServerError(c, "上传图片失败", err)
		return
	}

	response.Success(c, "上传图片成功", gin.H{
		"id":  postImage.ID,
		"url": postImage.URL,
	})
}

// AssociateImageWithPost 将临时图片关联到动态
func (h *ImageHandler) AssociateImageWithPost(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	type AssociateRequest struct {
		ImageID uint `json:"image_id" binding:"required"`
		PostID  uint `json:"post_id" binding:"required"`
	}

	var req AssociateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	// 关联图片到动态
	err := h.imageService.AssociateImageWithPost(c.Request.Context(), req.ImageID, req.PostID, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "关联图片失败", err)
		return
	}

	response.Success(c, "关联图片成功", nil)
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
