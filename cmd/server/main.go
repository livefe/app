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
	"app/internal/routes"
	"app/pkg/database"
	"app/pkg/logger"
	"app/pkg/redis"
	"app/pkg/validation"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	err := config.Init()
	if err != nil {
		fmt.Printf("配置初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 获取配置信息
	cfg := config.GetConfig()

	// 初始化日志组件
	err = logger.Init()
	if err != nil {
		fmt.Printf("日志初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化数据库连接
	err = database.InitGormDB()
	if err != nil {
		logger.Fatalf("数据库初始化失败: %v", err)
	}

	// 初始化Redis连接
	err = redis.InitRedis()
	if err != nil {
		logger.Fatalf("Redis初始化失败: %v", err)
	}

	// 初始化Gin引擎
	router := gin.Default()

	// 初始化验证器
	err = validation.Init()
	if err != nil {
		logger.Fatalf("验证器初始化失败: %v", err)
	}

	// 设置路由
	routes.SetupRouter(router)

	// 准备服务器地址
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// 启动HTTP服务器（非阻塞）
	go func() {
		logger.Infof("HTTP服务器正在启动，监听地址: %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 注册优雅关闭函数
	setupGracefulShutdown(srv)
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown(srv *http.Server) {
	// 创建一个接收系统信号的通道
	quit := make(chan os.Signal, 1)
	// 监听系统信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-quit
	logger.Info("正在关闭服务器...")

	// 创建一个超时上下文，等待现有请求完成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 停止接受新的HTTP请求
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("服务器关闭异常: %v", err)
	}
	logger.Info("HTTP服务已停止接受新请求")

	// 关闭数据库连接
	if err := database.CloseGormDB(); err != nil {
		logger.Errorf("关闭数据库连接失败: %v", err)
	}

	// 关闭Redis连接
	if err := redis.Close(); err != nil {
		logger.Errorf("关闭Redis连接失败: %v", err)
	}

	// 同步日志缓冲区
	logger.Sync()

	logger.Info("服务器已安全关闭")
	os.Exit(0)
}
