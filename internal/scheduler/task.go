package scheduler

import (
	"app/pkg/scheduler"
	"time"
)

// TaskConfig 定义任务配置结构
type TaskConfig struct {
	Spec        string                // Cron表达式
	Description string                // 任务描述
	Timeout     time.Duration         // 任务超时时间
	RetryCount  int                   // 失败重试次数
	Priority    int                   // 任务优先级（1-10，10为最高）
	Handler     scheduler.TaskHandler // 任务处理函数
}

// 定义所有定时任务的配置
var TaskConfigs = map[string]TaskConfig{
	"user_cleanup": {
		Spec:        "0 0 2 * * *", // 每天凌晨2点执行
		Description: "清理长时间未活跃的用户数据，包括临时文件和过期会话",
		Timeout:     30 * time.Minute,
		RetryCount:  3,
		Priority:    5,
		Handler:     UserCleanupTask,
	},
	"system_health": {
		Spec:        "0 */30 * * * *", // 每30分钟执行一次
		Description: "检查系统各组件的健康状态，包括数据库、缓存和外部服务连接",
		Timeout:     5 * time.Minute,
		RetryCount:  2,
		Priority:    8,
		Handler:     SystemHealthCheckTask,
	},
	"data_statistics": {
		Spec:        "0 0 1 * * *", // 每天凌晨1点执行
		Description: "生成系统数据统计报告，包括用户活跃度和系统资源使用情况",
		Timeout:     60 * time.Minute,
		RetryCount:  2,
		Priority:    4,
		Handler:     DataStatisticsTask,
	},
}

// RegisterTasks 注册所有定时任务
func RegisterTasks(s *scheduler.Scheduler) error {
	// 自动遍历TaskConfigs注册所有任务
	for taskName, config := range TaskConfigs {
		// 注册任务，直接使用配置中的处理函数
		err := s.Register(taskName, config.Spec, config.Handler)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTaskHandler 根据任务名称获取任务处理函数
func GetTaskHandler(taskName string) (scheduler.TaskHandler, bool) {
	// 从任务配置中获取处理函数
	config, exists := TaskConfigs[taskName]
	if !exists {
		return nil, false
	}

	return config.Handler, true
}

// GetTaskConfig 根据任务名称获取任务配置
func GetTaskConfig(taskName string) (TaskConfig, bool) {
	config, exists := TaskConfigs[taskName]
	return config, exists
}
