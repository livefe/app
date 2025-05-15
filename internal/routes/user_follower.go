// 粉丝关注相关路由定义
// 包含关注、取消关注、获取粉丝列表和关注列表等功能的API路由
package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterUserFollowerRoutes 注册粉丝关注相关路由
// 配置粉丝关注模块的所有API路由
// 参数:
//   - r: Gin路由引擎实例
func RegisterUserFollowerRoutes(r *gin.Engine) {
	// 从容器获取粉丝关注服务
	container := container.GetInstance()
	followerService := container.GetUserFollowerService()

	// 初始化粉丝关注处理器
	followerHandler := handler.NewUserFollowerHandler(followerService)

	// API根路径
	apiGroup := r.Group("/api")

	// 粉丝关注相关API组
	followerGroup := apiGroup.Group("/user_follower")

	// 注册需要认证的粉丝关注路由
	registerFollowerAuthRoutes(followerGroup, followerHandler)
}

// registerFollowerAuthRoutes 注册需要认证的粉丝关注相关路由
// 参数:
//   - group: 路由组
//   - handler: 粉丝关注处理器
func registerFollowerAuthRoutes(group *gin.RouterGroup, handler *handler.UserFollowerHandler) {
	// 添加认证中间件
	authGroup := group.Group("/", middleware.AuthMiddleware())

	// 关注操作
	authGroup.POST("/follow", handler.FollowUser)     // 关注用户
	authGroup.POST("/unfollow", handler.UnfollowUser) // 取消关注

	// 关系查询
	authGroup.GET("/followers/:user_id", handler.GetFollowers) // 获取粉丝列表
	authGroup.GET("/following/:user_id", handler.GetFollowing) // 获取关注列表
}
