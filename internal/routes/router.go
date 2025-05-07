package routes

import (
	"app/internal/container"
	"app/internal/middleware"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(r *gin.Engine) *gin.Engine {
	// 应用全局中间件
	r.Use(middleware.Logger())

	// 预初始化容器（确保所有依赖项都已准备好）
	_ = container.GetInstance()

	// 健康检查路由
	r.GET("/health", HealthCheck)

	// 注册用户路由
	RegisterUserRoutes(r)

	// 注册动态路由
	RegisterPostRoutes(r)

	// 注册粉丝关注相关路由
	RegisterFollowerRoutes(r)

	// 注册好友关系相关路由
	RegisterFriendRoutes(r)

	// 注意：已移除RegisterRelationRoutes，相关功能已由粉丝关注和好友关系模块替代

	return r
}

// HealthCheck 健康检查处理函数
func HealthCheck(c *gin.Context) {
	// 使用统一响应格式
	response.Success(c, "服务运行正常", gin.H{"status": "ok"})
}
