package routes

import (
	"app/internal/container"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterImageRoutes 注册图片相关路由
func RegisterImageRoutes(router *gin.Engine) {
	// 从容器获取图片处理器
	container := container.GetInstance()
	imageHandler := container.GetImageHandler()

	// 图片相关路由组
	imageGroup := router.Group("/api/images")
	{
		// 需要认证的路由
		authGroup := imageGroup.Group("/")
		authGroup.Use(middleware.AuthMiddleware())
		{
			// 上传临时图片（二进制文件方式，不关联动态ID）
			authGroup.POST("/temp", imageHandler.UploadTempImage)
			// 批量上传临时图片（多文件上传）
			authGroup.POST("/temp/multiple", imageHandler.UploadMultipleTempImages)
			// 上传用户头像（二进制文件方式）
			authGroup.POST("/avatar", imageHandler.UploadAvatar)
		}
	}
}
