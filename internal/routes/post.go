package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPostRoutes 注册动态相关路由
func RegisterPostRoutes(r *gin.Engine) {
	// 从容器获取动态服务
	container := container.GetInstance()
	postService := container.GetPostService()

	// 初始化处理器
	postHandler := handler.NewPostHandler(postService)

	// API根路径
	apiGroup := r.Group("/api")

	// 动态相关API组
	postGroup := apiGroup.Group("/social")

	// 添加认证中间件
	authGroup := postGroup.Group("/", middleware.AuthMiddleware())

	// 动态相关路由
	authGroup.POST("/post", postHandler.CreatePost)                   // 创建动态
	authGroup.GET("/posts", postHandler.GetPosts)                     // 获取动态列表
	authGroup.POST("/post/like", postHandler.LikePost)                // 点赞动态
	authGroup.POST("/post/comment", postHandler.CommentPost)          // 评论动态
	authGroup.GET("/post/comments/:post_id", postHandler.GetComments) // 获取评论列表
}
