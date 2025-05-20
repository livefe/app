// Package redis 提供Redis数据库操作的封装
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"app/config"

	"github.com/redis/go-redis/v9"
)

// Client 全局Redis客户端实例
var Client *redis.Client

// 错误常量
var (
	// ErrNil 表示Redis中不存在该键
	ErrNil = redis.Nil
	// ErrKeyNotFound 表示键不存在
	ErrKeyNotFound = errors.New("key not found")
	// ErrInvalidType 表示类型无效
	ErrInvalidType = errors.New("invalid type")
)

// 默认上下文超时时间
const defaultTimeout = 5 * time.Second

// RedisConfig Redis连接配置结构体
type RedisConfig struct {
	Addr         string        // Redis服务器地址
	Password     string        // Redis密码
	DB           int           // 数据库索引
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	DialTimeout  time.Duration // 连接超时时间
	ReadTimeout  time.Duration // 读取超时时间
	WriteTimeout time.Duration // 写入超时时间
}

// Init 初始化Redis连接并测试连接可用性
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

	// 测试连接
	ctx, cancel := getContext()
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		_ = client.Close()
		return fmt.Errorf("Redis连接测试失败: %w", err)
	}

	// 设置全局Client实例
	Client = client

	return nil
}

// parseRedisConfig 解析Redis配置
func parseRedisConfig() (*RedisConfig, error) {
	cfg := config.GetRedisConfig()
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// 解析超时时间配置
	dialTimeout, _ := time.ParseDuration(cfg.DialTimeout)
	readTimeout, _ := time.ParseDuration(cfg.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(cfg.WriteTimeout)

	return &RedisConfig{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}, nil
}

// Close 安全地关闭Redis连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// getContext 创建带默认超时的上下文
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

// 字符串操作

// Set 设置键值对并指定过期时间
func Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := getContext()
	defer cancel()
	return Client.Set(ctx, key, value, expiration).Err()
}

// Get 获取字符串类型的键值
func Get(key string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	result, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return result, nil
}

// GetObj 获取JSON对象并反序列化到指定结构
func GetObj(key string, obj interface{}) error {
	ctx, cancel := getContext()
	defer cancel()

	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrKeyNotFound
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), obj)
}

// SetObj 设置对象（序列化后存储）
func SetObj(key string, obj interface{}, expiration time.Duration) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	ctx, cancel := getContext()
	defer cancel()

	return Client.Set(ctx, key, data, expiration).Err()
}

// Del 删除键
func Del(keys ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Del(ctx, keys...).Result()
}

// Exists 检查键是否存在
func Exists(keys ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.Expire(ctx, key, expiration).Result()
}

// 哈希表操作

// HSet 设置哈希表字段
func HSet(key string, values ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.HSet(ctx, key, values...).Result()
}

// HGet 获取哈希表字段
func HGet(key, field string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	val, err := Client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// HGetAll 获取哈希表所有字段和值
func HGetAll(key string) (map[string]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希表字段
func HDel(key string, fields ...string) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.HDel(ctx, key, fields...).Result()
}

// 列表操作

// LPush 将一个或多个值插入到列表头部
func LPush(key string, values ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.LPush(ctx, key, values...).Result()
}

// RPush 将一个或多个值插入到列表尾部
func RPush(key string, values ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.RPush(ctx, key, values...).Result()
}

// LPop 移出并获取列表的第一个元素
func LPop(key string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	val, err := Client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// RPop 移出并获取列表的最后一个元素
func RPop(key string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	val, err := Client.RPop(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// LRange 获取列表指定范围内的元素
func LRange(key string, start, stop int64) ([]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.LRange(ctx, key, start, stop).Result()
}

// 集合操作

// SAdd 向集合添加一个或多个成员
func SAdd(key string, members ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.SAdd(ctx, key, members...).Result()
}

// SMembers 获取集合所有成员
func SMembers(key string) ([]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.SMembers(ctx, key).Result()
}

// SRem 移除集合中一个或多个成员
func SRem(key string, members ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.SRem(ctx, key, members...).Result()
}

// 有序集合操作

// ZAdd 向有序集合添加一个或多个成员
func ZAdd(key string, members ...redis.Z) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ZAdd(ctx, key, members...).Result()
}

// ZRange 通过索引区间返回有序集合成员
func ZRange(key string, start, stop int64) ([]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ZRange(ctx, key, start, stop).Result()
}

// ZRem 移除有序集合中的一个或多个成员
func ZRem(key string, members ...interface{}) (int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	return Client.ZRem(ctx, key, members...).Result()
}
