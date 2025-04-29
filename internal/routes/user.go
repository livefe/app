package routes

import (
	"app/internal/container"
	"app/internal/handler"
	"app/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(r *gin.Engine) {
	// 从容器获取用户服务
	container := container.GetInstance()
	userService := container.GetUserService()

	// 初始化处理器
	userHandler := handler.NewUserHandler(userService)

	// 用户API根路径
	apiGroup := r.Group("/api")

	// 用户相关API组
	userGroup := apiGroup.Group("/user")

	// 注册公开路由（无需认证）
	registerPublicRoutes(userGroup, userHandler)

	// 注册需要认证的路由
	registerAuthenticatedRoutes(userGroup, userHandler)
}

// registerPublicRoutes 注册公开路由（无需认证）
func registerPublicRoutes(group *gin.RouterGroup, handler *handler.UserHandler) {
	// 发送验证码
	group.POST("/verification-code", handler.SendVerificationCode)
	// 验证码登录
	group.POST("/login/code", handler.VerificationCodeLogin)
}

// registerAuthenticatedRoutes 注册需要认证的路由
func registerAuthenticatedRoutes(group *gin.RouterGroup, handler *handler.UserHandler) {
	// 添加认证中间件
	authGroup := group.Group("/", middleware.AuthMiddleware())

	// 获取用户信息
	authGroup.GET("/:id", handler.GetUserInfo)
	// 注销账号
	authGroup.POST("/deactivate", handler.DeactivateAccount)
	// 退出登录
	authGroup.POST("/logout", handler.Logout)
}
