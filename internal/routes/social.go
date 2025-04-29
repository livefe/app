package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterSocialRoutes 注册社交相关路由
func RegisterSocialRoutes(r *gin.Engine) {
	// 从容器获取社交服务
	container := container.GetInstance()
	socialService := container.GetSocialService()

	// 初始化处理器
	socialHandler := handler.NewSocialHandler(socialService)

	// 社交API根路径
	apiGroup := r.Group("/api")

	// 社交相关API组
	socialGroup := apiGroup.Group("/social")

	// 添加认证中间件
	authGroup := socialGroup.Group("/", middleware.AuthMiddleware())

	// 社交关系相关路由
	authGroup.POST("/follow", socialHandler.FollowUser)              // 关注用户
	authGroup.POST("/unfollow", socialHandler.UnfollowUser)          // 取消关注
	authGroup.GET("/followers/:user_id", socialHandler.GetFollowers) // 获取粉丝列表
	authGroup.GET("/following/:user_id", socialHandler.GetFollowing) // 获取关注列表

	// 位置分享相关路由
	authGroup.POST("/location", socialHandler.ShareLocation) // 分享位置
	authGroup.POST("/nearby", socialHandler.GetNearbyUsers)  // 获取附近用户

	// 社交动态相关路由
	authGroup.POST("/post", socialHandler.CreatePost)                   // 创建动态
	authGroup.GET("/posts", socialHandler.GetPosts)                     // 获取动态列表
	authGroup.POST("/post/like", socialHandler.LikePost)                // 点赞动态
	authGroup.POST("/post/comment", socialHandler.CommentPost)          // 评论动态
	authGroup.GET("/post/comments/:post_id", socialHandler.GetComments) // 获取评论列表
}
