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

// GetSocialRelationRepository 获取社交关系仓库实例（懒加载）
func (c *Container) GetSocialRelationRepository() repository.SocialRelationRepository {
	repo := c.getOrCreateRepository("social_relation_repository", func() interface{} {
		return repository.NewSocialRelationRepository(c.db)
	})
	return repo.(repository.SocialRelationRepository)
}

// GetSocialPostRepository 获取社交动态仓库实例（懒加载）
func (c *Container) GetSocialPostRepository() repository.SocialPostRepository {
	repo := c.getOrCreateRepository("social_post_repository", func() interface{} {
		return repository.NewSocialPostRepository(c.db)
	})
	return repo.(repository.SocialPostRepository)
}

// GetSocialCommentRepository 获取社交评论仓库实例（懒加载）
func (c *Container) GetSocialCommentRepository() repository.SocialCommentRepository {
	repo := c.getOrCreateRepository("social_comment_repository", func() interface{} {
		return repository.NewSocialCommentRepository(c.db)
	})
	return repo.(repository.SocialCommentRepository)
}

// GetSocialLocationShareRepository 获取位置分享仓库实例（懒加载）
func (c *Container) GetSocialLocationShareRepository() repository.SocialLocationShareRepository {
	repo := c.getOrCreateRepository("social_location_repository", func() interface{} {
		return repository.NewSocialLocationShareRepository(c.db)
	})
	return repo.(repository.SocialLocationShareRepository)
}

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

// GetSocialService 获取社交服务实例（懒加载）
func (c *Container) GetSocialService() service.SocialService {
	svc := c.getOrCreateService("social_service", func() interface{} {
		// 先获取依赖的仓库
		userRepo := c.GetUserRepository()
		relationRepo := c.GetSocialRelationRepository()
		locationRepo := c.GetSocialLocationShareRepository()
		postRepo := c.GetSocialPostRepository()
		commentRepo := c.GetSocialCommentRepository()

		return service.NewSocialService(
			relationRepo,
			locationRepo,
			postRepo,
			commentRepo,
			userRepo,
		)
	})
	return svc.(service.SocialService)
}
