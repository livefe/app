// 社交动态相关路由定义
// 包含动态发布、点赞、评论等功能的API路由
package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPostRoutes 注册社交动态相关路由
// 配置社交动态模块的所有API路由
// 参数:
//   - r: Gin路由引擎实例
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
// 参数:
//   - group: 路由组
//   - handler: 动态处理器
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
