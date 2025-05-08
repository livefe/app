package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterUserFollowerRoutes 注册粉丝关注相关路由
func RegisterUserFollowerRoutes(r *gin.Engine) {
	// 从容器获取粉丝关注服务
	container := container.GetInstance()
	followerService := container.GetUserFollowerService()

	// 初始化处理器
	followerHandler := handler.NewUserFollowerHandler(followerService)

	// API根路径
	apiGroup := r.Group("/api")

	// 粉丝关注相关API组
	followerGroup := apiGroup.Group("/user_follower")

	// 添加认证中间件
	authGroup := followerGroup.Group("/", middleware.AuthMiddleware())

	// 粉丝关注相关路由
	authGroup.POST("/follow", followerHandler.FollowUser)              // 关注用户
	authGroup.POST("/unfollow", followerHandler.UnfollowUser)          // 取消关注
	authGroup.GET("/followers/:user_id", followerHandler.GetFollowers) // 获取粉丝列表
	authGroup.GET("/following/:user_id", followerHandler.GetFollowing) // 获取关注列表
}
