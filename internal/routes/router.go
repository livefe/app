// Package routes 提供API路由注册和管理功能
// 负责组织和注册所有HTTP路由，包括公共路由和需要认证的路由
// 遵循模块化设计，每个功能模块的路由在单独的文件中定义
package routes

import (
	"app/internal/container"
	"app/internal/middleware"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置并注册所有API路由
// 应用全局中间件，注册各功能模块的路由，并返回配置好的路由引擎
// 参数:
//   - r: Gin路由引擎实例
//
// 返回:
//   - 配置完成的Gin路由引擎实例
func SetupRouter(r *gin.Engine) *gin.Engine {
	// 应用全局中间件
	r.Use(middleware.Logger())

	// 预初始化容器（确保所有依赖项都已准备好）
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

	// 粉丝关注模块路由
	RegisterUserFollowerRoutes(r)

	// 好友关系模块路由
	RegisterUserFriendRoutes(r)
}

// HealthCheck 处理健康检查请求
// 返回服务运行状态信息
// 参数:
//   - c: Gin上下文
func HealthCheck(c *gin.Context) {
	response.Success(c, "服务运行正常", gin.H{"status": "ok"})
}
