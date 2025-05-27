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
	// 从容器获取服务
	container := container.GetInstance()
	postHandler := container.GetPostHandler()

	// API根路径
	apiGroup := r.Group("/api")

	// 动态相关API组
	postGroup := apiGroup.Group("/post")

	// 注册需要认证的动态路由
	registerPostAuthRoutes(postGroup, postHandler)
}

// registerPostAuthRoutes 注册需要认证的动态相关路由
func registerPostAuthRoutes(group *gin.RouterGroup, postHandler *handler.PostHandler) {
	// 添加认证中间件
	authGroup := group.Group("/", middleware.AuthMiddleware())
	authGroup.POST("/create", postHandler.CreatePost)            // 创建动态
	authGroup.GET("/list", postHandler.GetPosts)                 // 获取动态列表
	authGroup.POST("/like", postHandler.LikePost)                // 点赞动态
	authGroup.POST("/comment", postHandler.CommentPost)          // 评论动态
	authGroup.GET("/comments/:post_id", postHandler.GetComments) // 获取评论列表
}
