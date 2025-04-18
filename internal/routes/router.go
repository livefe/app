package routes

import (
	"app/internal/middleware"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(r *gin.Engine) *gin.Engine {
	// 添加日志中间件
	r.Use(middleware.LoggerMiddleware())

	// 健康检查路由
	r.GET("/health", HealthCheck)

	// 注册用户路由
	RegisterUserRoutes(r)

	return r
}

// HealthCheck 健康检查处理函数
func HealthCheck(c *gin.Context) {
	// 使用统一响应格式
	response.Success(c, "服务运行正常", gin.H{"status": "ok"})
}
