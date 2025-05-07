package service

import (
	"app/internal/dto"
	"app/internal/model"
	"app/internal/repository"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// FollowerService 粉丝关注服务接口
type FollowerService interface {
	// 查询方法
	// GetFollowers 获取粉丝列表
	GetFollowers(ctx context.Context, req *dto.GetFollowersRequest) (*dto.GetFollowersResponse, error)
	// GetFollowing 获取关注列表
	GetFollowing(ctx context.Context, req *dto.GetFollowingRequest) (*dto.GetFollowingResponse, error)

	// 修改方法
	// FollowUser 关注用户
	FollowUser(ctx context.Context, req *dto.FollowUserRequest, userID uint) (*dto.FollowUserResponse, error)
	// UnfollowUser 取消关注用户
	UnfollowUser(ctx context.Context, req *dto.UnfollowUserRequest, userID uint) error
}

// followerService 粉丝关注服务实现
type followerService struct {
	followerRepo repository.FollowerRepository
	userRepo     repository.UserRepository
}

// NewFollowerService 创建粉丝关注服务实例
func NewFollowerService(
	followerRepo repository.FollowerRepository,
	userRepo repository.UserRepository,
) FollowerService {
	return &followerService{
		followerRepo: followerRepo,
		userRepo:     userRepo,
	}
}

// 查询方法

// GetFollowers 获取粉丝列表
func (s *followerService) GetFollowers(ctx context.Context, req *dto.GetFollowersRequest) (*dto.GetFollowersResponse, error) {
	// 获取粉丝关系列表
	followers, count, err := s.followerRepo.GetFollowers(req.UserID, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("获取粉丝列表失败: %w", err)
	}

	// 构建用户简要信息列表
	userList := make([]dto.UserBrief, 0, len(followers))
	for _, follower := range followers {
		user, err := s.userRepo.FindByID(follower.UserID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

		// 直接创建用户简要信息
		userList = append(userList, dto.UserBrief{
			ID:       user.ID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	return &dto.GetFollowersResponse{
		Total: int(count),
		List:  userList,
	}, nil
}

// GetFollowing 获取关注列表
func (s *followerService) GetFollowing(ctx context.Context, req *dto.GetFollowingRequest) (*dto.GetFollowingResponse, error) {
	// 获取关注关系列表
	following, count, err := s.followerRepo.GetFollowing(req.UserID, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("获取关注列表失败: %w", err)
	}

	// 构建用户简要信息列表
	userList := make([]dto.UserBrief, 0, len(following))
	for _, follow := range following {
		user, err := s.userRepo.FindByID(follow.TargetID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

		// 直接创建用户简要信息
		userList = append(userList, dto.UserBrief{
			ID:       user.ID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	return &dto.GetFollowingResponse{
		Total: int(count),
		List:  userList,
	}, nil
}

// 修改方法

// FollowUser 关注用户
func (s *followerService) FollowUser(ctx context.Context, req *dto.FollowUserRequest, userID uint) (*dto.FollowUserResponse, error) {
	// 检查目标用户是否存在
	_, err := s.userRepo.FindByID(req.TargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("目标用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查是否已关注
	_, err = s.followerRepo.GetFollower(userID, req.TargetID)
	if err == nil {
		return nil, errors.New("已经关注该用户")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询关注关系失败: %w", err)
	}

	// 创建关注关系
	follower := &model.Follower{
		UserID:   userID,
		TargetID: req.TargetID,
	}

	err = s.followerRepo.CreateFollower(follower)
	if err != nil {
		return nil, fmt.Errorf("创建关注关系失败: %w", err)
	}

	return &dto.FollowUserResponse{
		ID:        follower.ID,
		UserID:    follower.UserID,
		TargetID:  follower.TargetID,
		CreatedAt: follower.CreatedAt,
	}, nil
}

// UnfollowUser 取消关注用户
func (s *followerService) UnfollowUser(ctx context.Context, req *dto.UnfollowUserRequest, userID uint) error {
	// 检查是否已关注
	_, err := s.followerRepo.GetFollower(userID, req.TargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("未关注该用户")
		}
		return fmt.Errorf("查询关注关系失败: %w", err)
	}

	// 删除关注关系
	err = s.followerRepo.DeleteFollower(userID, req.TargetID)
	if err != nil {
		return fmt.Errorf("取消关注失败: %w", err)
	}

	return nil
}
