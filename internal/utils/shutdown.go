package utils

import (
	"app/pkg/database"
	"app/pkg/logger"
	"app/pkg/redis"
	"fmt"
)

// CloseResources 按照依赖关系的相反顺序关闭所有资源
// 确保资源释放的正确顺序，避免依赖问题
func CloseResources() {
	// 关闭数据库连接
	if err := database.Close(); err != nil {
		fmt.Printf("关闭数据库连接失败: %v\n", err)
	}

	// 关闭Redis连接
	if err := redis.Close(); err != nil {
		fmt.Printf("关闭Redis连接失败: %v\n", err)
	}

	// 关闭日志系统
	if err := logger.Close(); err != nil {
		fmt.Printf("关闭日志系统失败: %v\n", err)
	}
}