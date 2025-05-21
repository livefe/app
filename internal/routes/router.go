// Package routes 提供API路由注册和管理功能
package routes

import (
	"app/internal/container"
	"app/internal/middleware"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置并注册所有API路由
// 返回配置完成的Gin路由引擎实例
func SetupRouter(r *gin.Engine) *gin.Engine {
	// 应用全局中间件
	r.Use(middleware.Logger())

	// 预初始化容器
	_ = container.GetInstance()

	// 注册基础路由
	registerBaseRoutes(r)

	// 注册业务模块路由
	registerModuleRoutes(r)

	return r
}

// registerBaseRoutes 注册基础路由，如健康检查等不属于特定业务模块的路由
func registerBaseRoutes(r *gin.Engine) {
	// 健康检查路由
	r.GET("/health", HealthCheck)
}

// registerModuleRoutes 注册所有业务模块的路由
func registerModuleRoutes(r *gin.Engine) {
	// 用户模块路由
	RegisterUserRoutes(r)

	// 社交动态模块路由
	RegisterPostRoutes(r)

	// 用户关系模块路由
	RegisterRelationRoutes(r)
}

// HealthCheck 处理健康检查请求
func HealthCheck(c *gin.Context) {
	response.Success(c, "服务运行正常", gin.H{"status": "ok"})
}
