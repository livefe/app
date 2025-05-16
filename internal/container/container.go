package container

import (
	"app/internal/repository"
	"app/internal/service"
	"app/pkg/database"
	"sync"

	"gorm.io/gorm"
)

// Container 依赖注入容器，用于管理应用程序中的所有依赖项
type Container struct {
	// 数据库连接实例
	db *gorm.DB

	// 使用sync.Map替代互斥锁和字段，提高并发安全性
	repositories sync.Map
	services     sync.Map
}

// 全局容器实例
var instance *Container
var once sync.Once

// GetInstance 获取容器的单例实例
func GetInstance() *Container {
	once.Do(func() {
		instance = &Container{
			db: database.GetGormDB(),
		}
	})
	return instance
}

// 通用的获取或创建仓库实例的方法
func (c *Container) getOrCreateRepository(key string, creator func() interface{}) interface{} {
	if value, ok := c.repositories.Load(key); ok {
		return value
	}

	// 创建新实例
	repo := creator()

	// 使用LoadOrStore确保并发安全，即使有多个goroutine同时调用
	actual, _ := c.repositories.LoadOrStore(key, repo)
	return actual
}

// 通用的获取或创建服务实例的方法
func (c *Container) getOrCreateService(key string, creator func() interface{}) interface{} {
	if value, ok := c.services.Load(key); ok {
		return value
	}

	// 创建新实例
	svc := creator()

	// 使用LoadOrStore确保并发安全，即使有多个goroutine同时调用
	actual, _ := c.services.LoadOrStore(key, svc)
	return actual
}

// ==================== 仓库实例获取方法 ====================

// GetUserRepository 获取用户仓库实例（懒加载）
func (c *Container) GetUserRepository() repository.UserRepository {
	repo := c.getOrCreateRepository("user_repository", func() interface{} {
		return repository.NewUserRepository(c.db)
	})
	return repo.(repository.UserRepository)
}

// GetSMSRepository 获取短信仓库实例（懒加载）
func (c *Container) GetSMSRepository() repository.SMSRepository {
	repo := c.getOrCreateRepository("sms_repository", func() interface{} {
		return repository.NewSMSRepository(c.db)
	})
	return repo.(repository.SMSRepository)
}

// GetUserFollowerRepository 获取粉丝关注仓库实例（懒加载）
func (c *Container) GetUserFollowerRepository() repository.UserFollowerRepository {
	repo := c.getOrCreateRepository("user_follower_repository", func() interface{} {
		return repository.NewUserFollowerRepository(c.db)
	})
	return repo.(repository.UserFollowerRepository)
}

// GetUserFriendRepository 获取好友关系仓库实例（懒加载）
func (c *Container) GetUserFriendRepository() repository.UserFriendRepository {
	repo := c.getOrCreateRepository("user_friend_repository", func() interface{} {
		return repository.NewUserFriendRepository(c.db)
	})
	return repo.(repository.UserFriendRepository)
}

// GetPostRepository 获取动态仓库实例（懒加载）
func (c *Container) GetPostRepository() repository.PostRepository {
	repo := c.getOrCreateRepository("post_repository", func() interface{} {
		return repository.NewPostRepository(c.db)
	})
	return repo.(repository.PostRepository)
}

// GetPostCommentRepository 获取动态评论仓库实例（懒加载）
func (c *Container) GetPostCommentRepository() repository.PostCommentRepository {
	repo := c.getOrCreateRepository("post_comment_repository", func() interface{} {
		return repository.NewPostCommentRepository(c.db)
	})
	return repo.(repository.PostCommentRepository)
}

// ==================== 服务实例获取方法 ====================

// GetUserService 获取用户服务实例（懒加载）
func (c *Container) GetUserService() service.UserService {
	svc := c.getOrCreateService("user_service", func() interface{} {
		// 先获取依赖的仓库
		userRepo := c.GetUserRepository()
		smsRepo := c.GetSMSRepository()
		return service.NewUserService(userRepo, smsRepo)
	})
	return svc.(service.UserService)
}

// GetRelationService 获取用户关系服务实例（懒加载）
// 整合了粉丝关注和好友关系功能
func (c *Container) GetRelationService() service.RelationService {
	svc := c.getOrCreateService("relation_service", func() interface{} {
		return service.NewRelationService(
			c.GetUserFollowerRepository(),
			c.GetUserFriendRepository(),
			c.GetUserRepository(),
		)
	})
	return svc.(service.RelationService)
}

// 注意：以下旧的服务接口已被移除，请使用GetRelationService
// 如果需要这些服务，请在service包中定义相应接口

/*
// GetUserFollowerService 获取粉丝关注服务实例（懒加载）
func (c *Container) GetUserFollowerService() service.UserFollowerService {
	// 已废弃，请使用GetRelationService
	return nil
}

// GetUserFriendService 获取好友关系服务实例（懒加载）
func (c *Container) GetUserFriendService() service.UserFriendService {
	// 已废弃，请使用GetRelationService
	return nil
}
*/

// GetPostService 获取动态服务实例（懒加载）
func (c *Container) GetPostService() service.PostService {
	svc := c.getOrCreateService("post_service", func() interface{} {
		return service.NewPostService(
			c.GetPostRepository(),
			c.GetPostCommentRepository(),
			c.GetUserRepository(),
		)
	})
	return svc.(service.PostService)
}
