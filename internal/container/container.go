// Package container 提供依赖注入容器，管理应用程序中的所有依赖项
package container

import (
	"app/internal/repository"
	"app/internal/service"
	"app/pkg/database"
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// Container 依赖注入容器，管理应用程序中的服务和仓库实例
type Container struct {
	db           *gorm.DB // 数据库连接实例
	repositories sync.Map // 存储仓库实例的并发安全映射
	services     sync.Map // 存储服务实例的并发安全映射
}

var (
	instance *Container // 全局容器单例实例
	once     sync.Once  // 确保单例只被初始化一次
)

// GetInstance 返回容器的全局单例实例
// 线程安全，保证容器只被初始化一次
func GetInstance() *Container {
	once.Do(func() {
		instance = &Container{
			db: database.GetDB(),
		}
	})
	return instance
}

// getOrCreateRepository 获取已存在的仓库实例或创建新实例
// 使用懒加载模式，确保并发安全
func (c *Container) getOrCreateRepository(key string, creator func() interface{}) interface{} {
	if value, ok := c.repositories.Load(key); ok {
		return value
	}

	repo := creator()
	actual, _ := c.repositories.LoadOrStore(key, repo)
	return actual
}

// getOrCreateService 获取已存在的服务实例或创建新实例
// 使用懒加载模式，确保并发安全
func (c *Container) getOrCreateService(key string, creator func() interface{}) interface{} {
	if value, ok := c.services.Load(key); ok {
		return value
	}

	svc := creator()
	actual, _ := c.services.LoadOrStore(key, svc)
	return actual
}

// ==================== 仓库实例获取方法 ====================

// GetUserRepository 返回用户仓库实例
func (c *Container) GetUserRepository() repository.UserRepository {
	repo := c.getOrCreateRepository("user_repository", func() interface{} {
		return repository.NewUserRepository(c.db)
	})
	return repo.(repository.UserRepository)
}

// GetSMSRepository 返回短信仓库实例
func (c *Container) GetSMSRepository() repository.SMSRepository {
	repo := c.getOrCreateRepository("sms_repository", func() interface{} {
		return repository.NewSMSRepository(c.db)
	})
	return repo.(repository.SMSRepository)
}

// GetUserFollowerRepository 返回粉丝关注仓库实例
func (c *Container) GetUserFollowerRepository() repository.UserFollowerRepository {
	repo := c.getOrCreateRepository("user_follower_repository", func() interface{} {
		return repository.NewUserFollowerRepository(c.db)
	})
	return repo.(repository.UserFollowerRepository)
}

// GetUserFriendRepository 返回好友关系仓库实例
func (c *Container) GetUserFriendRepository() repository.UserFriendRepository {
	repo := c.getOrCreateRepository("user_friend_repository", func() interface{} {
		return repository.NewUserFriendRepository(c.db)
	})
	return repo.(repository.UserFriendRepository)
}

// GetPostRepository 返回动态仓库实例
func (c *Container) GetPostRepository() repository.PostRepository {
	repo := c.getOrCreateRepository("post_repository", func() interface{} {
		return repository.NewPostRepository(c.db)
	})
	return repo.(repository.PostRepository)
}

// GetPostCommentRepository 返回动态评论仓库实例
func (c *Container) GetPostCommentRepository() repository.PostCommentRepository {
	repo := c.getOrCreateRepository("post_comment_repository", func() interface{} {
		return repository.NewPostCommentRepository(c.db)
	})
	return repo.(repository.PostCommentRepository)
}

// GetPostImageRepository 返回动态图片仓库实例
func (c *Container) GetPostImageRepository() repository.PostImageRepository {
	repo := c.getOrCreateRepository("post_image_repository", func() interface{} {
		return repository.NewPostImageRepository(c.db)
	})
	return repo.(repository.PostImageRepository)
}

// ==================== 服务实例获取方法 ====================

// GetUserService 返回用户服务实例
func (c *Container) GetUserService() service.UserService {
	svc := c.getOrCreateService("user_service", func() interface{} {
		return service.NewUserService(
			c.GetUserRepository(),
			c.GetSMSRepository(),
			c.GetImageService(),
		)
	})
	return svc.(service.UserService)
}

// GetRelationService 返回用户关系服务实例
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

// GetPostService 返回动态服务实例
func (c *Container) GetPostService() service.PostService {
	svc := c.getOrCreateService("post_service", func() interface{} {
		return service.NewPostService(
			c.GetPostRepository(),
			c.GetPostCommentRepository(),
			c.GetUserRepository(),
			c.GetPostImageRepository(),
			c.GetImageService(),
		)
	})
	return svc.(service.PostService)
}

// GetImageService 返回图片服务实例
func (c *Container) GetImageService() service.ImageService {
	svc := c.getOrCreateService("image_service", func() interface{} {
		imageService, err := service.NewImageService(
			c.GetPostImageRepository(),
			c.GetUserRepository(),
		)
		if err != nil {
			panic(fmt.Sprintf("创建图片服务失败: %v", err))
		}
		return imageService
	})
	return svc.(service.ImageService)
}
