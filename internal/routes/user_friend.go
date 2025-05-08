package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterUserFriendRoutes 注册好友关系相关路由
func RegisterUserFriendRoutes(r *gin.Engine) {
	// 从容器获取好友关系服务
	container := container.GetInstance()
	friendService := container.GetUserFriendService()

	// 初始化处理器
	friendHandler := handler.NewUserFriendHandler(friendService)

	// API根路径
	apiGroup := r.Group("/api")

	// 好友关系相关API组
	friendGroup := apiGroup.Group("/user_friend")

	// 添加认证中间件
	authGroup := friendGroup.Group("/", middleware.AuthMiddleware())

	// 好友关系相关路由
	authGroup.POST("/add", friendHandler.AddFriend)             // 添加好友
	authGroup.POST("/accept", friendHandler.AcceptFriend)       // 接受好友请求
	authGroup.POST("/reject", friendHandler.RejectFriend)       // 拒绝好友请求
	authGroup.POST("/delete", friendHandler.DeleteFriend)       // 删除好友
	authGroup.GET("/requests", friendHandler.GetFriendRequests) // 获取好友请求列表
	authGroup.GET("/list", friendHandler.GetFriends)            // 获取好友列表
}
