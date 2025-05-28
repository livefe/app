package handler

import (
	"app/internal/service"
	"app/pkg/response"
	"io"
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

// UploadTempImage 上传临时图片
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

// UploadMultipleTempImages 批量上传临时图片
func (h *ImageHandler) UploadMultipleTempImages(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 获取上传的文件（多文件表单）
	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "获取上传文件失败", err)
		return
	}

	// 获取所有图片文件
	files := form.File["images"]
	if len(files) == 0 {
		response.BadRequest(c, "未找到上传的图片", nil)
		return
	}

	// 检查文件数量限制（最多10张）
	if len(files) > 10 {
		response.BadRequest(c, "一次最多上传10张图片", nil)
		return
	}

	// 准备参数
	filenames := make([]string, len(files))
	sizes := make([]int64, len(files))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	// 先检查所有文件的有效性
	for i, file := range files {
		// 检查文件大小
		if file.Size > 10*1024*1024 {
			response.BadRequest(c, "文件大小超过限制: "+file.Filename, nil)
			return
		}

		// 检查文件类型
		ext := filepath.Ext(file.Filename)
		if !validExts[ext] {
			response.BadRequest(c, "不支持的文件类型: "+file.Filename, nil)
			return
		}

		// 保存文件信息
		filenames[i] = file.Filename
		sizes[i] = file.Size
	}

	// 打开所有文件
	readers := make([]io.Reader, len(files))
	openedFiles := make([]io.ReadCloser, len(files))
	for i, file := range files {
		src, err := file.Open()
		if err != nil {
			// 关闭已经打开的文件
			for j := 0; j < i; j++ {
				openedFiles[j].Close()
			}
			response.InternalServerError(c, "打开上传文件失败: "+file.Filename, err)
			return
		}
		openedFiles[i] = src
		readers[i] = src
	}

	// 确保所有文件都会被关闭
	defer func() {
		for _, f := range openedFiles {
			if f != nil {
				f.Close()
			}
		}
	}()

	// 批量上传图片
	tempImages, errs := h.imageService.UploadMultipleTempImages(c.Request.Context(), userID.(uint), readers, filenames, sizes)

	// 检查是否全部失败
	if len(tempImages) == 0 {
		response.InternalServerError(c, "所有图片上传失败", errs[0])
		return
	}

	// 准备响应数据
	imagesData := make([]map[string]interface{}, 0, len(tempImages))
	for _, img := range tempImages {
		imagesData = append(imagesData, map[string]interface{}{
			"id":           img.ID,
			"url":          img.URL,
			"size":         img.Size,
			"content_type": img.ContentType,
			"filename":     filepath.Base(img.ObjectKey),
		})
	}

	// 统计上传结果
	successCount := len(tempImages)
	failCount := len(files) - successCount

	response.Success(c, "上传完成", gin.H{
		"total":         len(files),
		"success_count": successCount,
		"fail_count":    failCount,
		"images":        imagesData,
	})
}
