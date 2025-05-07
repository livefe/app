package container

import (
	"app/internal/repository"
	"app/internal/service"
)

// GetFollowerRepository 获取粉丝关注仓库实例（懒加载）
func (c *Container) GetFollowerRepository() repository.FollowerRepository {
	repo := c.getOrCreateRepository("follower_repository", func() interface{} {
		return repository.NewFollowerRepository(c.db)
	})
	return repo.(repository.FollowerRepository)
}

// GetFriendRepository 获取好友关系仓库实例（懒加载）
func (c *Container) GetFriendRepository() repository.FriendRepository {
	repo := c.getOrCreateRepository("friend_repository", func() interface{} {
		return repository.NewFriendRepository(c.db)
	})
	return repo.(repository.FriendRepository)
}

// GetFollowerService 获取粉丝关注服务实例（懒加载）
func (c *Container) GetFollowerService() service.FollowerService {
	svc := c.getOrCreateService("follower_service", func() interface{} {
		return service.NewFollowerService(
			c.GetFollowerRepository(),
			c.GetUserRepository(),
		)
	})
	return svc.(service.FollowerService)
}

// GetFriendService 获取好友关系服务实例（懒加载）
func (c *Container) GetFriendService() service.FriendService {
	svc := c.getOrCreateService("friend_service", func() interface{} {
		return service.NewFriendService(
			c.GetFriendRepository(),
			c.GetUserRepository(),
		)
	})
	return svc.(service.FriendService)
}
