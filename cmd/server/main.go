// Package main 实现API服务器的入口点
// 负责初始化配置、数据库连接、路由和HTTP服务器
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
	"app/internal/routes"
	"app/internal/utils"
	"app/pkg/database"
	"app/pkg/logger"
	"app/pkg/redis"
	"app/pkg/validation"

	"github.com/gin-gonic/gin"
)

// main 是API服务器的入口函数
// 按顺序初始化各个组件，启动HTTP服务器，并设置优雅关闭机制
func main() {
	// 初始化应用程序组件
	initComponents()

	// 获取配置信息
	cfg := config.GetConfig()

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

	// 初始化验证器
	if err := validation.Init(); err != nil {
		fmt.Printf("验证器初始化失败: %v\n", err)
		os.Exit(1)
	}
}

// setupHTTPServer 配置并启动HTTP服务器
// 返回服务器实例以便后续优雅关闭
func setupHTTPServer(cfg *config.Config) *http.Server {
	// 初始化Gin引擎
	router := gin.Default()

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
		fmt.Printf("HTTP服务器正在启动，监听地址: %s\n", serverAddr)
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
	fmt.Println("正在关闭服务器...")

	// 创建一个超时上下文，等待现有请求完成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 停止接受新的HTTP请求
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("服务器关闭异常: %v\n", err)
	}
	fmt.Println("HTTP服务已停止接受新请求")

	// 按照依赖关系的相反顺序关闭资源
	utils.CloseResources()

	fmt.Println("服务器已安全关闭")
	os.Exit(0)
}
