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

	// 仓库层实例
	userRepository           repository.UserRepository
	smsRepository            repository.SMSRepository
	socialRelationRepository repository.SocialRelationRepository
	socialPostRepository     repository.SocialPostRepository
	socialCommentRepository  repository.SocialCommentRepository
	socialLocationRepository repository.SocialLocationShareRepository

	// 服务层实例
	userService   service.UserService
	socialService service.SocialService

	// 确保单例模式的互斥锁
	mutex sync.Mutex
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

// GetUserRepository 获取用户仓库实例（懒加载）
func (c *Container) GetUserRepository() repository.UserRepository {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.userRepository == nil {
		c.userRepository = repository.NewUserRepository(c.db)
	}
	return c.userRepository
}

// GetSMSRepository 获取短信仓库实例（懒加载）
func (c *Container) GetSMSRepository() repository.SMSRepository {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.smsRepository == nil {
		c.smsRepository = repository.NewSMSRepository(c.db)
	}
	return c.smsRepository
}

// GetSocialRelationRepository 获取社交关系仓库实例（懒加载）
func (c *Container) GetSocialRelationRepository() repository.SocialRelationRepository {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.socialRelationRepository == nil {
		c.socialRelationRepository = repository.NewSocialRelationRepository(c.db)
	}
	return c.socialRelationRepository
}

// GetSocialPostRepository 获取社交动态仓库实例（懒加载）
func (c *Container) GetSocialPostRepository() repository.SocialPostRepository {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.socialPostRepository == nil {
		c.socialPostRepository = repository.NewSocialPostRepository(c.db)
	}
	return c.socialPostRepository
}

// GetSocialCommentRepository 获取社交评论仓库实例（懒加载）
func (c *Container) GetSocialCommentRepository() repository.SocialCommentRepository {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.socialCommentRepository == nil {
		c.socialCommentRepository = repository.NewSocialCommentRepository(c.db)
	}
	return c.socialCommentRepository
}

// GetSocialLocationShareRepository 获取位置分享仓库实例（懒加载）
func (c *Container) GetSocialLocationShareRepository() repository.SocialLocationShareRepository {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.socialLocationRepository == nil {
		c.socialLocationRepository = repository.NewSocialLocationShareRepository(c.db)
	}
	return c.socialLocationRepository
}

// GetUserService 获取用户服务实例（懒加载）
func (c *Container) GetUserService() service.UserService {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.userService == nil {
		userRepo := c.GetUserRepository()
		smsRepo := c.GetSMSRepository()
		c.userService = service.NewUserService(userRepo, smsRepo)
	}
	return c.userService
}

// GetSocialService 获取社交服务实例（懒加载）
func (c *Container) GetSocialService() service.SocialService {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.socialService == nil {
		// 使用新的社交服务实现，直接注入各个子仓库
		userRepo := c.GetUserRepository()
		relationRepo := c.GetSocialRelationRepository()
		locationRepo := c.GetSocialLocationShareRepository()
		postRepo := c.GetSocialPostRepository()
		commentRepo := c.GetSocialCommentRepository()

		// 使用正确的构造函数创建社交服务实例
		c.socialService = service.NewSocialService(
			relationRepo,
			locationRepo,
			postRepo,
			commentRepo,
			userRepo,
		)
	}
	return c.socialService
}
