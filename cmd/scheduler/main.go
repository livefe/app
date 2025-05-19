// Package main 实现定时任务调度服务的入口点
// 负责初始化配置、数据库连接、定时任务调度器和HTTP管理接口
// 并提供优雅关闭机制确保资源正确释放
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/config"
	"app/internal/scheduler"
	"app/pkg/database"
	"app/pkg/logger"
	"app/pkg/redis"
	pkgscheduler "app/pkg/scheduler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 全局定时任务调度器实例
var schedulerInstance *pkgscheduler.Scheduler

// main 是定时任务调度服务的入口函数
// 按顺序初始化各个组件，启动调度器和HTTP管理接口，并设置优雅关闭机制
func main() {
	// 初始化应用程序组件
	initComponents()

	// 获取配置信息
	cfg := config.GetConfig()

	// 初始化并启动调度器
	initAndStartScheduler()

	// 设置并启动HTTP服务器
	srv := setupHTTPServer(cfg)

	// 注册优雅关闭函数
	setupGracefulShutdown(srv)
}

// initComponents 按顺序初始化所有应用程序组件
// 任何组件初始化失败都会导致程序退出
func initComponents() {
	// 初始化配置
	if err := config.Init(); err != nil {
		fmt.Printf("配置初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化数据库连接
	if err := database.Init(); err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化Redis连接
	if err := redis.Init(); err != nil {
		fmt.Printf("Redis初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	if err := logger.Init(); err != nil {
		fmt.Printf("日志系统初始化失败: %v\n", err)
		os.Exit(1)
	}
}

// initAndStartScheduler 初始化并启动定时任务调度器
// 注册所有任务并启动调度器
func initAndStartScheduler() {
	// 初始化定时任务调度器（使用Redis分布式锁）
	schedulerInstance = pkgscheduler.Init(pkgscheduler.WithRedisLock())

	// 注册所有定时任务
	ctx := context.Background()
	for taskName, config := range scheduler.TaskConfigs {
		// 创建注册选项，使用任务配置中的设置
		options := pkgscheduler.RegisterOption{
			RunImmediately: config.RunImmediately, // 使用配置中的立即执行设置
			LockTimeout:    config.LockTimeout,    // 使用配置中的锁超时设置
		}

		// 使用选项注册任务
		err := schedulerInstance.RegisterWithOptions(taskName, config.Spec, config.Handler, options)
		if err != nil {
			logger.Error(ctx, "注册定时任务失败", zap.String("task", taskName), zap.Error(err))
			os.Exit(1)
		}
		logger.Info(ctx, "成功注册定时任务", zap.String("task", taskName), zap.String("spec", config.Spec))
	}

	// 启动定时任务调度器
	schedulerInstance.Start()
}

// setupHTTPServer 配置并启动HTTP服务器
// 返回服务器实例以便后续优雅关闭
func setupHTTPServer(cfg *config.Config) *http.Server {
	// 初始化Gin引擎
	router := gin.Default()

	// 设置API路由
	setupRouter(router)

	// 准备服务器地址
	serverAddr := fmt.Sprintf("%s:%d", cfg.Scheduler.Host, cfg.Scheduler.Port) // 使用Scheduler配置

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// 启动HTTP服务器（非阻塞）
	go func() {
		fmt.Printf("定时任务HTTP服务器正在启动，监听地址: %s\n", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务器启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	return srv
}

// setupGracefulShutdown 设置优雅关闭机制
// 监听系统信号，确保在关闭前完成所有请求并释放资源
func setupGracefulShutdown(srv *http.Server) {
	// 创建一个接收系统信号的通道
	quit := make(chan os.Signal, 1)
	// 监听系统信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-quit
	fmt.Println("正在关闭定时任务服务器...")

	// 停止定时任务调度器
	schedulerInstance.Stop()

	// 创建一个超时上下文，等待现有请求完成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 停止接受新的HTTP请求
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("服务器关闭异常: %v\n", err)
	}
	fmt.Println("HTTP服务已停止接受新请求")

	// 按照依赖关系的相反顺序关闭资源
	closeResources()

	fmt.Println("定时任务服务器已安全关闭")
	os.Exit(0)
}

// closeResources 按照依赖关系的相反顺序关闭所有资源
// 确保资源释放的正确顺序，避免依赖问题
func closeResources() {
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

// setupRouter 设置HTTP路由
// 配置健康检查和任务管理API接口
func setupRouter(router *gin.Engine) {
	// 健康检查接口
	router.GET("/health", handleHealthCheck)

	// 任务管理API组
	taskGroup := router.Group("/tasks")
	{
		// 获取所有任务列表
		taskGroup.GET("", handleGetAllTasks)

		// 获取指定任务信息
		taskGroup.GET("/:name", handleGetTaskInfo)

		// 手动执行任务
		taskGroup.POST("/:name/run", handleRunTask)
	}
}

// handleHealthCheck 处理健康检查请求
func handleHealthCheck(c *gin.Context) {
	if schedulerInstance.HealthCheck() {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "定时任务调度器运行正常",
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "定时任务调度器异常",
		})
	}
}

// handleGetAllTasks 处理获取所有任务列表请求
func handleGetAllTasks(c *gin.Context) {
	tasks := schedulerInstance.GetAllTasksInfo()
	c.JSON(http.StatusOK, tasks)
}

// handleGetTaskInfo 处理获取指定任务信息请求
func handleGetTaskInfo(c *gin.Context) {
	name := c.Param("name")
	taskInfo, err := schedulerInstance.GetTaskInfo(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, taskInfo)
}

// handleRunTask 处理手动执行任务请求
func handleRunTask(c *gin.Context) {
	name := c.Param("name")
	err := schedulerInstance.RunTask(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("任务 %s 已手动触发执行", name),
	})
}
