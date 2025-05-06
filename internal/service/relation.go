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

// RelationService 关系服务接口
type RelationService interface {
	// FollowUser 关注用户
	FollowUser(ctx context.Context, req *dto.FollowUserRequest, userID uint) (*dto.FollowUserResponse, error)
	// UnfollowUser 取消关注用户
	UnfollowUser(ctx context.Context, req *dto.UnfollowUserRequest, userID uint) error
	// GetFollowers 获取粉丝列表
	GetFollowers(ctx context.Context, req *dto.GetFollowersRequest) (*dto.GetFollowersResponse, error)
	// GetFollowing 获取关注列表
	GetFollowing(ctx context.Context, req *dto.GetFollowingRequest) (*dto.GetFollowingResponse, error)
}

// relationService 关系服务实现
type relationService struct {
	relationRepo repository.RelationRepository
	userRepo     repository.UserRepository
}

// NewRelationService 创建关系服务实例
func NewRelationService(
	relationRepo repository.RelationRepository,
	userRepo repository.UserRepository,
) RelationService {
	return &relationService{
		relationRepo: relationRepo,
		userRepo:     userRepo,
	}
}

// FollowUser 关注用户
func (s *relationService) FollowUser(ctx context.Context, req *dto.FollowUserRequest, userID uint) (*dto.FollowUserResponse, error) {
	// 检查目标用户是否存在
	_, err := s.userRepo.FindByID(req.TargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("目标用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查是否已关注
	_, err = s.relationRepo.GetRelation(userID, req.TargetID)
	if err == nil {
		return nil, errors.New("已经关注该用户")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询关系失败: %w", err)
	}

	// 创建关注关系
	relation := &model.Relation{
		UserID:   userID,
		TargetID: req.TargetID,
		Type:     1, // 关注类型
		Status:   1, // 已确认状态
	}

	err = s.relationRepo.CreateRelation(relation)
	if err != nil {
		return nil, fmt.Errorf("创建关注关系失败: %w", err)
	}

	return &dto.FollowUserResponse{
		ID:        relation.ID,
		UserID:    relation.UserID,
		TargetID:  relation.TargetID,
		CreatedAt: relation.CreatedAt,
	}, nil
}

// UnfollowUser 取消关注用户
func (s *relationService) UnfollowUser(ctx context.Context, req *dto.UnfollowUserRequest, userID uint) error {
	// 检查是否已关注
	_, err := s.relationRepo.GetRelation(userID, req.TargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("未关注该用户")
		}
		return fmt.Errorf("查询关系失败: %w", err)
	}

	// 删除关注关系
	err = s.relationRepo.DeleteRelation(userID, req.TargetID)
	if err != nil {
		return fmt.Errorf("取消关注失败: %w", err)
	}

	return nil
}

// GetFollowers 获取粉丝列表
func (s *relationService) GetFollowers(ctx context.Context, req *dto.GetFollowersRequest) (*dto.GetFollowersResponse, error) {
	// 获取粉丝关系列表
	relations, count, err := s.relationRepo.GetFollowers(req.UserID, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("获取粉丝列表失败: %w", err)
	}

	// 构建用户简要信息列表
	userList := make([]dto.UserBrief, 0, len(relations))
	for _, relation := range relations {
		user, err := s.userRepo.FindByID(relation.UserID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

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
func (s *relationService) GetFollowing(ctx context.Context, req *dto.GetFollowingRequest) (*dto.GetFollowingResponse, error) {
	// 获取关注关系列表
	relations, count, err := s.relationRepo.GetFollowing(req.UserID, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("获取关注列表失败: %w", err)
	}

	// 构建用户简要信息列表
	userList := make([]dto.UserBrief, 0, len(relations))
	for _, relation := range relations {
		user, err := s.userRepo.FindByID(relation.TargetID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

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
