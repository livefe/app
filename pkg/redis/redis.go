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

// RedisConfig Redis连接配置
type RedisConfig struct {
	Addr         string        // Redis服务器地址
	Password     string        // Redis密码
	DB           int           // Redis数据库
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	DialTimeout  time.Duration // 连接超时时间
	ReadTimeout  time.Duration // 读取超时时间
	WriteTimeout time.Duration // 写入超时时间
}

// InitRedis 初始化Redis连接
func InitRedis() error {
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

	// 使用Ping命令测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	if _, err := client.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("Redis连接测试失败: %w", err)
	}

	// 设置全局Client实例
	Client = client

	return nil
}

// parseRedisConfig 解析Redis配置
func parseRedisConfig() (*RedisConfig, error) {
	// 获取Redis配置
	cfg := config.GetRedisConfig()

	// 构建Redis地址
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return &RedisConfig{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: 10, // 默认最小空闲连接数
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}, nil
}

// Close 关闭Redis连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
