// 用户相关路由定义
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
	userHandler := container.GetUserHandler()

	// 用户API根路径
	apiGroup := r.Group("/api")

	// 用户相关API组
	userGroup := apiGroup.Group("/user")

	// 注册用户模块的路由
	registerUserPublicRoutes(userGroup, userHandler)
	registerUserAuthRoutes(userGroup, userHandler)
}

// registerUserPublicRoutes 注册用户模块的公开路由（无需认证）
func registerUserPublicRoutes(group *gin.RouterGroup, handler *handler.UserHandler) {
	// 认证相关
	group.POST("/verification-code", handler.SendVerificationCode) // 发送验证码
	group.POST("/login/code", handler.VerificationCodeLogin)       // 验证码登录
}

// registerUserAuthRoutes 注册用户模块的认证路由（需要认证）
func registerUserAuthRoutes(group *gin.RouterGroup, handler *handler.UserHandler) {
	// 添加认证中间件
	authGroup := group.Group("/", middleware.AuthMiddleware())

	// 用户信息管理
	authGroup.GET("/:id", handler.GetUserInfo)               // 获取用户信息
	authGroup.POST("/deactivate", handler.DeactivateAccount) // 注销账号
	authGroup.POST("/logout", handler.Logout)                // 退出登录
}
