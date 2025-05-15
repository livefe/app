// 好友关系相关路由定义
// 包含添加好友、接受/拒绝好友请求、删除好友、获取好友列表等功能的API路由
package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterUserFriendRoutes 注册好友关系相关路由
// 配置好友关系模块的所有API路由
// 参数:
//   - r: Gin路由引擎实例
func RegisterUserFriendRoutes(r *gin.Engine) {
	// 从容器获取好友关系服务
	container := container.GetInstance()
	friendService := container.GetUserFriendService()

	// 初始化好友关系处理器
	friendHandler := handler.NewUserFriendHandler(friendService)

	// API根路径
	apiGroup := r.Group("/api")

	// 好友关系相关API组
	friendGroup := apiGroup.Group("/user_friend")

	// 注册需要认证的好友关系路由
	registerFriendAuthRoutes(friendGroup, friendHandler)
}

// registerFriendAuthRoutes 注册需要认证的好友关系相关路由
// 参数:
//   - group: 路由组
//   - handler: 好友关系处理器
func registerFriendAuthRoutes(group *gin.RouterGroup, handler *handler.UserFriendHandler) {
	// 添加认证中间件
	authGroup := group.Group("/", middleware.AuthMiddleware())

	// 好友请求管理
	authGroup.POST("/add", handler.AddFriend)       // 添加好友
	authGroup.POST("/accept", handler.AcceptFriend) // 接受好友请求
	authGroup.POST("/reject", handler.RejectFriend) // 拒绝好友请求
	authGroup.POST("/delete", handler.DeleteFriend) // 删除好友

	// 好友信息查询
	authGroup.GET("/requests", handler.GetFriendRequests) // 获取好友请求列表
	authGroup.GET("/list", handler.GetFriends)            // 获取好友列表
}
