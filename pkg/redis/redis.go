// Package redis 提供Redis数据库操作的封装，包括连接管理和常用操作
package redis

import (
	"context"
	"fmt"
	"time"

	"app/config"

	"github.com/redis/go-redis/v9"
)

// Client 全局Redis客户端实例
var Client *redis.Client

// RedisConfig Redis连接配置结构体
type RedisConfig struct {
	Addr         string        // Redis服务器地址，格式为host:port
	Password     string        // Redis密码，如无密码则为空字符串
	DB           int           // Redis数据库索引，默认为0
	PoolSize     int           // 连接池大小，影响并发性能
	MinIdleConns int           // 最小空闲连接数，保持连接池活跃
	DialTimeout  time.Duration // 连接超时时间，建立连接的最大等待时间
	ReadTimeout  time.Duration // 读取超时时间，读取操作的最大等待时间
	WriteTimeout time.Duration // 写入超时时间，写入操作的最大等待时间
}

// Init 初始化Redis连接并测试连接可用性
// 成功时返回nil，失败时返回带上下文的错误信息
func Init() error {
	// 获取并解析Redis配置
	redisConfig, err := parseRedisConfig()
	if err != nil {
		return fmt.Errorf("解析Redis配置失败: %w", err)
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:         redisConfig.Addr,
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
		DialTimeout:  redisConfig.DialTimeout,
		ReadTimeout:  redisConfig.ReadTimeout,
		WriteTimeout: redisConfig.WriteTimeout,
	})

	// 使用Ping命令测试连接，设置5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	if _, err := client.Ping(ctx).Result(); err != nil {
		// 关闭客户端，避免资源泄漏
		_ = client.Close()
		return fmt.Errorf("Redis连接测试失败: %w", err)
	}

	// 设置全局Client实例
	Client = client

	return nil
}

// parseRedisConfig 解析Redis配置，将配置文件中的设置转换为Redis客户端所需的配置结构
// 返回RedisConfig结构体指针和可能的错误
func parseRedisConfig() (*RedisConfig, error) {
	// 获取Redis配置
	cfg := config.GetRedisConfig()

	// 构建Redis地址
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// 返回配置，设置合理的默认值
	return &RedisConfig{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: 10,               // 默认最小空闲连接数
		DialTimeout:  5 * time.Second,  // 连接超时5秒
		ReadTimeout:  3 * time.Second,  // 读取超时3秒
		WriteTimeout: 3 * time.Second,  // 写入超时3秒
	}, nil
}

// Close 安全地关闭Redis连接，释放资源
// 如果客户端不存在，则直接返回nil
// 返回关闭过程中可能发生的错误
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
