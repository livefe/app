package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"app/pkg/logger"
	"app/pkg/redis"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron      *cron.Cron
	entryMap  map[string]cron.EntryID
	handlers  map[string]TaskHandler
	redisLock bool // 是否使用Redis分布式锁
	mu        sync.RWMutex
}

// TaskHandler 任务处理函数类型
type TaskHandler func(ctx context.Context) error

// TaskInfo 任务信息
type TaskInfo struct {
	Name     string    // 任务名称
	Spec     string    // cron表达式
	Next     time.Time // 下次执行时间
	Prev     time.Time // 上次执行时间
	Running  bool      // 是否正在运行
	Disabled bool      // 是否禁用
}

// Init 初始化并返回一个新的调度器
func Init(opts ...Option) *Scheduler {
	// 创建带有秒级精度的cron调度器
	c := cron.New(cron.WithSeconds())

	s := &Scheduler{
		cron:      c,
		entryMap:  make(map[string]cron.EntryID),
		handlers:  make(map[string]TaskHandler),
		redisLock: false,
	}

	// 应用选项
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Option 调度器选项
type Option func(*Scheduler)

// WithRedisLock 启用Redis分布式锁
func WithRedisLock() Option {
	return func(s *Scheduler) {
		s.redisLock = true
	}
}

// Register 注册定时任务
func (s *Scheduler) Register(name, spec string, handler TaskHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查任务是否已存在
	if _, exists := s.entryMap[name]; exists {
		return fmt.Errorf("任务 %s 已存在", name)
	}

	// 包装处理函数，添加日志和错误处理
	wrappedHandler := func() {
		ctx := context.Background()
		logger.Info(ctx, "开始执行定时任务", zap.String("task", name))

		// 如果启用了Redis分布式锁，尝试获取锁
		if s.redisLock {
			lockKey := fmt.Sprintf("scheduler:lock:%s", name)
			// 设置锁过期时间为5分钟，防止任务执行时间过长导致锁永久存在
			success, err := s.acquireLock(ctx, lockKey, 5*time.Minute)
			if err != nil {
				logger.Error(ctx, "获取分布式锁失败", zap.String("task", name), zap.Error(err))
				return
			}
			if !success {
				logger.Info(ctx, "任务正在其他节点执行，跳过", zap.String("task", name))
				return
			}
			defer s.releaseLock(ctx, lockKey)
		}

		// 执行任务
		start := time.Now()
		err := handler(ctx)
		elapsed := time.Since(start)

		if err != nil {
			logger.Error(ctx, "定时任务执行失败",
				zap.String("task", name),
				zap.Duration("elapsed", elapsed),
				zap.Error(err))
		} else {
			logger.Info(ctx, "定时任务执行成功",
				zap.String("task", name),
				zap.Duration("elapsed", elapsed))
		}
	}

	// 添加到cron
	entryID, err := s.cron.AddFunc(spec, wrappedHandler)
	if err != nil {
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	// 保存任务信息
	s.entryMap[name] = entryID
	s.handlers[name] = handler

	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
	logger.Info(context.Background(), "定时任务调度器已启动")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cron.Stop()
	logger.Info(context.Background(), "定时任务调度器已停止")
}

// Remove 移除定时任务
func (s *Scheduler) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entryMap[name]; exists {
		s.cron.Remove(entryID)
		delete(s.entryMap, name)
		delete(s.handlers, name)
		logger.Info(context.Background(), "定时任务已移除", zap.String("task", name))
	}
}

// RunTask 手动执行定时任务
func (s *Scheduler) RunTask(name string) error {
	s.mu.RLock()
	handler, exists := s.handlers[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("任务 %s 不存在", name)
	}

	go func() {
		ctx := context.Background()
		logger.Info(ctx, "手动执行定时任务", zap.String("task", name))

		start := time.Now()
		err := handler(ctx)
		elapsed := time.Since(start)

		if err != nil {
			logger.Error(ctx, "手动执行定时任务失败",
				zap.String("task", name),
				zap.Duration("elapsed", elapsed),
				zap.Error(err))
		} else {
			logger.Info(ctx, "手动执行定时任务成功",
				zap.String("task", name),
				zap.Duration("elapsed", elapsed))
		}
	}()

	return nil
}

// GetTaskInfo 获取任务信息
func (s *Scheduler) GetTaskInfo(name string) (*TaskInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entryID, exists := s.entryMap[name]
	if !exists {
		return nil, fmt.Errorf("任务 %s 不存在", name)
	}

	entry := s.cron.Entry(entryID)
	return &TaskInfo{
		Name:     name,
		Spec:     "", // cron库不提供获取spec的方法
		Next:     entry.Next,
		Prev:     entry.Prev,
		Running:  false, // cron库不提供获取运行状态的方法
		Disabled: false, // cron库不提供获取禁用状态的方法
	}, nil
}

// ListTasks 列出所有任务
func (s *Scheduler) ListTasks() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]string, 0, len(s.entryMap))
	for name := range s.entryMap {
		tasks = append(tasks, name)
	}

	return tasks
}

// GetAllTasksInfo 获取所有任务信息
func (s *Scheduler) GetAllTasksInfo() map[string]TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]TaskInfo)
	for name, entryID := range s.entryMap {
		entry := s.cron.Entry(entryID)
		result[name] = TaskInfo{
			Name:     name,
			Spec:     "", // cron库不提供获取spec的方法
			Next:     entry.Next,
			Prev:     entry.Prev,
			Running:  false, // cron库不提供获取运行状态的方法
			Disabled: false, // cron库不提供获取禁用状态的方法
		}
	}

	return result
}

// HealthCheck 健康检查
func (s *Scheduler) HealthCheck() bool {
	return s.cron != nil
}

// 获取Redis分布式锁
func (s *Scheduler) acquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	redisClient := redis.Client
	if redisClient == nil {
		return false, fmt.Errorf("Redis客户端未初始化")
	}

	// 尝试获取锁，使用SET NX命令
	success, err := redisClient.SetNX(ctx, key, "1", expiration).Result()
	if err != nil {
		return false, fmt.Errorf("获取Redis锁失败: %w", err)
	}

	return success, nil
}

// 释放Redis分布式锁
func (s *Scheduler) releaseLock(ctx context.Context, key string) {
	redisClient := redis.Client
	if redisClient == nil {
		logger.Error(ctx, "释放锁失败: Redis客户端未初始化", zap.String("key", key))
		return
	}

	// 删除锁
	_, err := redisClient.Del(ctx, key).Result()
	if err != nil {
		logger.Error(ctx, "释放Redis锁失败", zap.String("key", key), zap.Error(err))
	}
}
