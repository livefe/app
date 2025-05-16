// 用户关系相关路由定义
// 包含关注、好友等社交关系功能的API路由
package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRelationRoutes 注册用户关系相关路由
// 配置用户关系模块的所有API路由，包括粉丝关注和好友关系
// 参数:
//   - r: Gin路由引擎实例
func RegisterRelationRoutes(r *gin.Engine) {
	// 从容器获取用户关系服务
	container := container.GetInstance()
	relationService := container.GetRelationService()

	// 初始化用户关系处理器
	relationHandler := handler.NewRelationHandler(relationService)

	// API根路径
	apiGroup := r.Group("/api")

	// 用户关系相关API组
	relationGroup := apiGroup.Group("/relation")

	// 注册需要认证的用户关系路由
	registerRelationAuthRoutes(relationGroup, relationHandler)
}

// registerRelationAuthRoutes 注册需要认证的用户关系相关路由
// 参数:
//   - group: 路由组
//   - handler: 用户关系处理器
func registerRelationAuthRoutes(group *gin.RouterGroup, handler *handler.RelationHandler) {
	// 添加认证中间件
	authGroup := group.Group("/", middleware.AuthMiddleware())

	// 关注操作
	authGroup.POST("/follow", handler.FollowUser)     // 关注用户
	authGroup.POST("/unfollow", handler.UnfollowUser) // 取消关注

	// 关系查询
	authGroup.GET("/followers/:user_id", handler.GetFollowers) // 获取粉丝列表
	authGroup.GET("/following/:user_id", handler.GetFollowing) // 获取关注列表

	// 好友请求管理
	authGroup.POST("/friend/add", handler.AddFriend)       // 添加好友
	authGroup.POST("/friend/accept", handler.AcceptFriend) // 接受好友请求
	authGroup.POST("/friend/reject", handler.RejectFriend) // 拒绝好友请求
	authGroup.POST("/friend/delete", handler.DeleteFriend) // 删除好友

	// 好友信息查询
	authGroup.GET("/friend/requests", handler.GetFriendRequests) // 获取好友请求列表
	authGroup.GET("/friend/list", handler.GetFriends)            // 获取好友列表
}
