// 社交动态相关路由定义
package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPostRoutes 注册社交动态相关路由
func RegisterPostRoutes(r *gin.Engine) {
	// 从容器获取动态服务
	container := container.GetInstance()
	postService := container.GetPostService()

	// 初始化动态处理器
	postHandler := handler.NewPostHandler(postService)

	// API根路径
	apiGroup := r.Group("/api")

	// 动态相关API组
	postGroup := apiGroup.Group("/post")

	// 注册需要认证的动态路由
	registerPostAuthRoutes(postGroup, postHandler)
}

// registerPostAuthRoutes 注册需要认证的动态相关路由
func registerPostAuthRoutes(group *gin.RouterGroup, handler *handler.PostHandler) {
	// 添加认证中间件
	authGroup := group.Group("/", middleware.AuthMiddleware())

	// 动态管理
	authGroup.POST("/create", handler.CreatePost) // 创建动态
	authGroup.GET("/list", handler.GetPosts)      // 获取动态列表

	// 互动功能
	authGroup.POST("/like", handler.LikePost)                // 点赞动态
	authGroup.POST("/comment", handler.CommentPost)          // 评论动态
	authGroup.GET("/comments/:post_id", handler.GetComments) // 获取评论列表
}
