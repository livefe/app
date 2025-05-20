// Package redis 提供Redis数据库操作的封装
package redis

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// 分布式锁相关错误
var (
	// ErrLockAcquireFailed 表示获取锁失败
	ErrLockAcquireFailed = errors.New("获取锁失败")
	// ErrLockNotHeld 表示当前未持有锁
	ErrLockNotHeld = errors.New("当前未持有锁")
)

// DistributedLock 分布式锁结构体
type DistributedLock struct {
	key        string        // 锁的键名
	value      string        // 锁的值（用于标识锁的持有者）
	expiration time.Duration // 锁的过期时间
}

// NewLock 创建一个新的分布式锁
func NewLock(key string, expiration time.Duration) *DistributedLock {
	return &DistributedLock{
		key:        key,
		value:      uuid.New().String(), // 使用UUID作为锁的值，确保唯一性
		expiration: expiration,
	}
}

// Acquire 获取锁，如果获取失败则返回错误
func (dl *DistributedLock) Acquire() error {
	ctx, cancel := getContext()
	defer cancel()

	// 使用SetNX尝试获取锁
	success, err := Client.SetNX(ctx, dl.key, dl.value, dl.expiration).Result()
	if err != nil {
		return err
	}

	if !success {
		return ErrLockAcquireFailed
	}

	return nil
}

// Release 释放锁，确保只有锁的持有者才能释放锁
func (dl *DistributedLock) Release() error {
	ctx, cancel := getContext()
	defer cancel()

	// Lua脚本，确保只有锁的持有者才能释放锁
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
	`

	// 执行Lua脚本
	result, err := Client.Eval(ctx, script, []string{dl.key}, dl.value).Int64()
	if err != nil {
		return err
	}

	if result == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// TryAcquire 尝试获取锁，如果获取失败则立即返回false
func (dl *DistributedLock) TryAcquire() (bool, error) {
	ctx, cancel := getContext()
	defer cancel()

	// 使用SetNX尝试获取锁
	return Client.SetNX(ctx, dl.key, dl.value, dl.expiration).Result()
}
