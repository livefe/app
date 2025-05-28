// 图片上传相关路由定义
package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterImageRoutes 注册图片相关路由
func RegisterImageRoutes(r *gin.Engine) {
	// 从容器获取图片处理器
	container := container.GetInstance()
	imageHandler := container.GetImageHandler()

	// 图片相关路由组
	imageGroup := r.Group("/api/images")

	// 注册需要认证的图片路由
	registerImageAuthRoutes(imageGroup, imageHandler)
}

// registerImageAuthRoutes 注册需要认证的图片相关路由
func registerImageAuthRoutes(group *gin.RouterGroup, handler *handler.ImageHandler) {
	// 添加认证中间件
	authGroup := group.Group("/", middleware.AuthMiddleware())

	authGroup.POST("/temp", handler.UploadTempImage)                   // 上传临时图片
	authGroup.POST("/temp/multiple", handler.UploadMultipleTempImages) // 批量上传临时图片
}
