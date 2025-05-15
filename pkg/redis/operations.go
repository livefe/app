// Package redis 提供Redis数据库操作的封装
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// 定义包级错误常量，便于统一错误处理
var (
	// ErrNil 表示Redis中不存在该键，与redis.Nil保持一致
	ErrNil = redis.Nil
	// ErrKeyNotFound 表示键不存在，用于统一API返回
	ErrKeyNotFound = errors.New("key not found")
	// ErrInvalidType 表示类型无效，用于类型转换失败时
	ErrInvalidType = errors.New("invalid type")
)

// 默认上下文超时时间
const defaultTimeout = 5 * time.Second

// getContext 创建带默认超时的上下文
// 返回上下文和取消函数，调用方负责在适当时机调用cancel
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

// executeWithRetry 执行带重试的Redis操作，最多重试3次
// 参数operation为要执行的Redis操作函数
// 返回操作结果错误，如果是预期的错误(如键不存在)则直接返回不重试
func executeWithRetry(operation func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		err = operation()
		// 如果操作成功或是预期的错误(键不存在)，直接返回
		if err == nil || err == redis.Nil || err == ErrKeyNotFound {
			return err
		}

		// 如果是连接相关错误，则等待后重试
		if isConnectionError(err) {
			// 指数退避策略，每次重试等待时间增加
			time.Sleep(time.Duration(i*100) * time.Millisecond)
			continue
		}

		// 其他错误直接返回，不重试
		return err
	}
	return err // 达到最大重试次数后返回最后一次错误
}

// isConnectionError 判断是否为网络连接相关错误
// 通过检查错误信息中的关键词来识别连接问题
// 返回布尔值表示是否是连接错误
func isConnectionError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "refused") ||
		strings.Contains(errStr, "network")
}

// ======== 字符串操作 ========

// Set 设置键值对并指定过期时间
// 参数:
//   - key: Redis键名
//   - value: 要存储的值，可以是任意类型
//   - expiration: 过期时间，0表示永不过期
//
// 返回可能的错误
func Set(key string, value interface{}, expiration time.Duration) error {
	return executeWithRetry(func() error {
		ctx, cancel := getContext()
		defer cancel()
		return Client.Set(ctx, key, value, expiration).Err()
	})
}

// Get 获取字符串类型的键值
// 参数:
//   - key: Redis键名
//
// 返回:
//   - 字符串值
//   - 可能的错误，键不存在时返回ErrKeyNotFound
func Get(key string) (string, error) {
	var val string
	err := executeWithRetry(func() error {
		ctx, cancel := getContext()
		defer cancel()

		result, err := Client.Get(ctx, key).Result()
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		if err != nil {
			return err
		}
		val = result
		return nil
	})

	return val, err
}

// GetObj 获取JSON对象并反序列化到指定结构
// 参数:
//   - key: Redis键名
//   - obj: 接收反序列化结果的对象指针
//
// 返回可能的错误，键不存在时返回ErrKeyNotFound
func GetObj(key string, obj interface{}) error {
	return executeWithRetry(func() error {
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
	})
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
	var count int64
	err := executeWithRetry(func() error {
		ctx, cancel := getContext()
		defer cancel()

		result, err := Client.Del(ctx, keys...).Result()
		if err != nil {
			return err
		}
		count = result
		return nil
	})

	return count, err
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

// ======== 哈希表操作 ========

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

// ======== 列表操作 ========

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

// ======== 集合操作 ========

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

// ======== 有序集合操作 ========

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
