package scheduler

import (
	"context"
	"fmt"
	"math/rand"
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
	// 创建带有秒级精度的cron调度器，并设置不立即执行任务
	c := cron.New(cron.WithSeconds(), cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),  // 如果上一次任务还在运行，则跳过本次执行
		cron.DelayIfStillRunning(cron.DefaultLogger), // 如果上一次任务还在运行，则延迟到上一次任务完成后执行
	))

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

	// 如果启用了Redis分布式锁，启动时检查并清理可能存在的死锁
	// 添加延迟，避免启动时立即清理导致的锁冲突
	if s.redisLock {
		go func() {
			// 延迟5秒后再清理死锁，避免多个实例同时启动时的冲突
			time.Sleep(5 * time.Second)
			s.cleanupDeadLocks()
		}()
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

// RegisterOption 注册任务的选项
type RegisterOption struct {
	RunImmediately bool          // 是否在添加后立即执行一次
	LockTimeout    time.Duration // 分布式锁超时时间
}

// DefaultRegisterOption 默认注册选项
var DefaultRegisterOption = RegisterOption{
	RunImmediately: true,
	LockTimeout:    5 * time.Minute,
}

// Register 注册定时任务
func (s *Scheduler) Register(name, spec string, handler TaskHandler) error {
	return s.RegisterWithOptions(name, spec, handler, DefaultRegisterOption)
}

// RegisterWithOptions 使用自定义选项注册定时任务
func (s *Scheduler) RegisterWithOptions(name, spec string, handler TaskHandler, options RegisterOption) error {
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
			// 使用选项中指定的锁超时时间，或默认值
			lockExpiration := options.LockTimeout
			if lockExpiration <= 0 {
				lockExpiration = 5 * time.Minute
			}

			// 创建分布式锁
			lock := redis.NewLock(lockKey, lockExpiration)

			// 尝试获取锁，添加随机延迟避免多个实例同时竞争
			randDelay := time.Duration(rand.Intn(500)) * time.Millisecond
			time.Sleep(randDelay)

			// 尝试获取锁
			success, err := lock.TryAcquire()
			if err != nil {
				logger.Error(ctx, "获取分布式锁失败", zap.String("task", name), zap.Error(err))
				return
			}
			if !success {
				logger.Info(ctx, "任务正在其他节点执行，跳过", zap.String("task", name))
				return
			}
			// 使用defer释放锁
			defer func() {
				if err := lock.Release(); err != nil {
					logger.Error(ctx, "释放分布式锁失败", zap.String("task", name), zap.Error(err))
				} else {
					logger.Debug(ctx, "成功释放分布式锁", zap.String("task", name))
				}
			}()
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
	var entryID cron.EntryID

	if !options.RunImmediately {
		// 使用Schedule方法而不是AddFunc，可以控制是否立即执行
		// 使用cron.Parse而不是cron.ParseStandard以支持秒级精度
		parser := cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		)
		schedule, err := parser.Parse(spec)
		if err != nil {
			return fmt.Errorf("解析cron表达式失败: %w", err)
		}
		entryID = s.cron.Schedule(schedule, cron.FuncJob(wrappedHandler))
	} else {
		// 使用自定义解析器确保支持秒级精度，并在添加后立即执行一次
		parser := cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		)
		schedule, err := parser.Parse(spec)
		if err != nil {
			return fmt.Errorf("解析cron表达式失败: %w", err)
		}
		// 添加任务并立即执行一次
		entryID = s.cron.Schedule(schedule, cron.FuncJob(wrappedHandler))
		go wrappedHandler() // 立即执行一次
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

// cleanupDeadLocks 检查并清理可能存在的死锁
func (s *Scheduler) cleanupDeadLocks() {
	ctx := context.Background()
	redisClient := redis.Client
	if redisClient == nil {
		logger.Error(ctx, "清理死锁失败: Redis客户端未初始化")
		return
	}

	// 获取所有任务的锁键
	s.mu.RLock()
	var lockKeys []string
	for name := range s.handlers {
		lockKeys = append(lockKeys, fmt.Sprintf("scheduler:lock:%s", name))
	}
	s.mu.RUnlock()

	// 检查并清理每个任务的锁
	for _, key := range lockKeys {
		// 检查锁是否存在
		exists, err := redisClient.Exists(ctx, key).Result()
		if err != nil {
			logger.Error(ctx, "检查锁状态失败", zap.String("key", key), zap.Error(err))
			continue
		}

		// 如果锁存在，获取锁的值（时间戳）
		if exists > 0 {
			// 获取锁的剩余过期时间
			ttl, err := redisClient.TTL(ctx, key).Result()
			if err != nil {
				logger.Error(ctx, "获取锁过期时间失败", zap.String("key", key), zap.Error(err))
				continue
			}

			// 如果锁没有设置过期时间或过期时间过长，则清理它
			if ttl < 0 || ttl > 30*time.Minute {
				_, err := redisClient.Del(ctx, key).Result()
				if err != nil {
					logger.Error(ctx, "清理死锁失败", zap.String("key", key), zap.Error(err))
				} else {
					logger.Info(ctx, "成功清理死锁", zap.String("key", key), zap.Duration("ttl", ttl))
				}
			} else {
				logger.Info(ctx, "发现有效锁，保留不清理", zap.String("key", key), zap.Duration("ttl", ttl))
			}
		}
	}
}
