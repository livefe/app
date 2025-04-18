package routes

import (
	"app/internal/handler"
	"app/internal/middleware"
	"app/internal/repository"
	"app/internal/service"
	"app/pkg/database"

	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(r *gin.Engine) {
	// 获取数据库连接
	db := database.GetGormDB()

	// 初始化仓库
	userRepo := repository.NewUserRepository(db)
	// 初始化SMS记录仓库
	smsRepo := repository.NewSMSRepository(db)

	// 初始化服务
	userService := service.NewUserService(userRepo, smsRepo)

	// 初始化处理器
	userHandler := handler.NewUserHandler(userService)

	// 用户相关API组
	userGroup := r.Group("/api/user")
	{
		// 发送验证码
		userGroup.POST("/verification-code", userHandler.SendVerificationCode)
		// 验证码登录
		userGroup.POST("/login/code", userHandler.VerificationCodeLogin)
		// 获取用户信息 - 添加JWT认证中间件
		userGroup.GET("/:id", middleware.AuthMiddleware(), userHandler.GetUserInfo)
	}
}
