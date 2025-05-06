package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRelationRoutes 注册关系相关路由
func RegisterRelationRoutes(r *gin.Engine) {
	// 从容器获取关系服务
	container := container.GetInstance()
	relationService := container.GetRelationService()

	// 初始化处理器
	relationHandler := handler.NewRelationHandler(relationService)

	// API根路径
	apiGroup := r.Group("/api")

	// 关系相关API组
	relationGroup := apiGroup.Group("/relation")

	// 添加认证中间件
	authGroup := relationGroup.Group("/", middleware.AuthMiddleware())

	// 关系相关路由
	authGroup.POST("/follow", relationHandler.FollowUser)              // 关注用户
	authGroup.POST("/unfollow", relationHandler.UnfollowUser)          // 取消关注
	authGroup.GET("/followers/:user_id", relationHandler.GetFollowers) // 获取粉丝列表
	authGroup.GET("/following/:user_id", relationHandler.GetFollowing) // 获取关注列表
}
