package service

import (
	"app/internal/constant"
	"app/internal/dto"
	"app/internal/model"
	"app/internal/repository"
	"context"
	"errors"

	"gorm.io/gorm"
)

// RelationService 用户关系服务接口
type RelationService interface {
	// FollowUser 关注用户
	FollowUser(ctx context.Context, req *dto.FollowUserRequest, userID uint) (*dto.FollowUserResponse, error)
	// UnfollowUser 取消关注用户
	UnfollowUser(ctx context.Context, req *dto.UnfollowUserRequest, userID uint) error
	// GetFollowers 获取粉丝列表
	GetFollowers(ctx context.Context, req *dto.GetFollowersRequest) (*dto.GetFollowersResponse, error)
	// GetFollowing 获取关注列表
	GetFollowing(ctx context.Context, req *dto.GetFollowingRequest) (*dto.GetFollowingResponse, error)
	// AddFriend 添加好友
	AddFriend(ctx context.Context, req *dto.AddFriendRequest, userID uint) (*dto.AddFriendResponse, error)
	// AcceptFriend 接受好友请求
	AcceptFriend(ctx context.Context, req *dto.AcceptFriendRequest, userID uint) error
	// RejectFriend 拒绝好友请求
	RejectFriend(ctx context.Context, req *dto.RejectFriendRequest, userID uint) error
	// DeleteFriend 删除好友
	DeleteFriend(ctx context.Context, req *dto.DeleteFriendRequest, userID uint) error
	// GetFriendRequests 获取好友请求列表
	GetFriendRequests(ctx context.Context, req *dto.GetFriendRequestsRequest, userID uint) (*dto.GetFriendRequestsResponse, error)
	// GetFriends 获取好友列表
	GetFriends(ctx context.Context, req *dto.GetFriendsRequest, userID uint) (*dto.GetFriendsResponse, error)
}

// relationService 用户关系服务实现
type relationService struct {
	followerRepo repository.UserFollowerRepository
	friendRepo   repository.UserFriendRepository
	userRepo     repository.UserRepository
}

// NewRelationService 创建用户关系服务实例
func NewRelationService(
	followerRepo repository.UserFollowerRepository,
	friendRepo repository.UserFriendRepository,
	userRepo repository.UserRepository,
) RelationService {
	return &relationService{
		followerRepo: followerRepo,
		friendRepo:   friendRepo,
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
		return nil, err
	}

	// 检查是否已关注
	existingFollower, err := s.followerRepo.GetFollower(userID, req.TargetID)
	exists := err == nil && existingFollower != nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if exists {
		return nil, errors.New("已经关注该用户")
	}

	// 创建关注关系
	newFollower := &model.UserFollower{
		UserID:   userID,
		TargetID: req.TargetID,
	}

	// 保存到数据库
	err = s.followerRepo.CreateFollower(newFollower)
	if err != nil {
		return nil, err
	}

	return &dto.FollowUserResponse{
		ID:        newFollower.ID,
		UserID:    newFollower.UserID,
		TargetID:  newFollower.TargetID,
		CreatedAt: newFollower.CreatedAt,
	}, nil
}

// UnfollowUser 取消关注用户
func (s *relationService) UnfollowUser(ctx context.Context, req *dto.UnfollowUserRequest, userID uint) error {
	// 检查是否已关注
	follower, err := s.followerRepo.GetFollower(userID, req.TargetID)
	exists := err == nil && follower != nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if !exists {
		return errors.New("未关注该用户")
	}

	// 删除关注关系
	return s.followerRepo.DeleteFollower(userID, req.TargetID)
}

// GetFollowers 获取粉丝列表
func (s *relationService) GetFollowers(ctx context.Context, req *dto.GetFollowersRequest) (*dto.GetFollowersResponse, error) {
	// 获取粉丝关系列表
	followers, total, err := s.followerRepo.GetFollowers(req.UserID, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	// 构建响应数据
	list := make([]dto.UserBrief, 0, len(followers))
	for _, follower := range followers {
		// 获取粉丝用户信息
		user, err := s.userRepo.FindByID(follower.UserID)
		if err != nil {
			continue
		}

		// 添加到列表
		list = append(list, dto.UserBrief{
			ID:       user.ID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	return &dto.GetFollowersResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetFollowing 获取关注列表
func (s *relationService) GetFollowing(ctx context.Context, req *dto.GetFollowingRequest) (*dto.GetFollowingResponse, error) {
	// 获取关注关系列表
	followings, total, err := s.followerRepo.GetFollowing(req.UserID, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	// 构建响应数据
	list := make([]dto.UserBrief, 0, len(followings))
	for _, following := range followings {
		// 获取关注用户信息
		user, err := s.userRepo.FindByID(following.TargetID)
		if err != nil {
			continue
		}

		// 添加到列表
		list = append(list, dto.UserBrief{
			ID:       user.ID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	return &dto.GetFollowingResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// AddFriend 添加好友
func (s *relationService) AddFriend(ctx context.Context, req *dto.AddFriendRequest, userID uint) (*dto.AddFriendResponse, error) {
	// 检查目标用户是否存在
	_, err := s.userRepo.FindByID(req.TargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("目标用户不存在")
		}
		return nil, err
	}

	// 检查是否已经是好友
	friend, err := s.friendRepo.GetFriend(userID, req.TargetID)
	isFriend := err == nil && friend != nil
	if err != nil {
		return nil, err
	}
	if isFriend {
		return nil, errors.New("已经是好友关系")
	}

	// 检查是否已经发送过好友请求
	// 前面已经获取过friend，如果不是好友关系，检查是否有待处理的请求
	if friend != nil && friend.Status == int(constant.FriendStatusPending) {
		return nil, errors.New("已经发送过好友请求")
	}

	// 创建好友请求
	friendRequest := &model.UserFriend{
		UserID:    userID,
		TargetID:  req.TargetID,
		Status:    int(constant.FriendStatusPending),
		Direction: 0, // 发起方
	}

	// 保存到数据库
	err = s.friendRepo.CreateFriend(friendRequest)
	if err != nil {
		return nil, err
	}

	return &dto.AddFriendResponse{
		ID:        friendRequest.ID,
		UserID:    friendRequest.UserID,
		TargetID:  friendRequest.TargetID,
		Status:    friendRequest.Status,
		CreatedAt: friendRequest.CreatedAt,
	}, nil
}

// AcceptFriend 接受好友请求
func (s *relationService) AcceptFriend(ctx context.Context, req *dto.AcceptFriendRequest, userID uint) error {
	// 获取好友请求
	friendRequest, err := s.friendRepo.GetFriendByID(req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友请求不存在")
		}
		return err
	}

	// 检查请求是否发给当前用户
	if friendRequest.TargetID != userID {
		return errors.New("无权操作此好友请求")
	}

	// 检查请求状态
	if friendRequest.Status != int(constant.FriendStatusPending) {
		return errors.New("好友请求已处理")
	}

	// 更新好友请求状态为已接受
	return s.friendRepo.UpdateFriendStatus(friendRequest.ID, int(constant.FriendStatusConfirmed))
}

// RejectFriend 拒绝好友请求
func (s *relationService) RejectFriend(ctx context.Context, req *dto.RejectFriendRequest, userID uint) error {
	// 获取好友请求
	friendRequest, err := s.friendRepo.GetFriendByID(req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友请求不存在")
		}
		return err
	}

	// 检查请求是否发给当前用户
	if friendRequest.TargetID != userID {
		return errors.New("无权操作此好友请求")
	}

	// 检查请求状态
	if friendRequest.Status != int(constant.FriendStatusPending) {
		return errors.New("好友请求已处理")
	}

	// 更新请求状态为已拒绝
	return s.friendRepo.UpdateFriendStatus(friendRequest.ID, 2) // 拒绝状态值为2
}

// DeleteFriend 删除好友
func (s *relationService) DeleteFriend(ctx context.Context, req *dto.DeleteFriendRequest, userID uint) error {
	// 检查是否是好友关系
	friend, err := s.friendRepo.GetFriend(userID, req.TargetID)
	isFriend := err == nil && friend != nil && friend.Status == int(constant.FriendStatusConfirmed)
	if err != nil {
		return err
	}
	if !isFriend {
		return errors.New("不是好友关系")
	}

	// 删除好友关系（双向）
	return s.friendRepo.DeleteFriend(userID, req.TargetID)
}

// GetFriendRequests 获取好友请求列表
func (s *relationService) GetFriendRequests(ctx context.Context, req *dto.GetFriendRequestsRequest, userID uint) (*dto.GetFriendRequestsResponse, error) {
	// 获取好友请求列表
	requests, total, err := s.friendRepo.GetFriendRequests(userID, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	// 构建响应数据
	list := make([]dto.FriendRequestItem, 0, len(requests))
	for _, request := range requests {
		// 获取请求用户信息
		user, err := s.userRepo.FindByID(request.UserID)
		if err != nil {
			continue
		}

		// 添加到列表
		list = append(list, dto.FriendRequestItem{
			ID:        request.ID,
			UserID:    user.ID,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Message:   "", // Message字段在请求中不存在，使用空字符串
			CreatedAt: request.CreatedAt,
		})
	}

	return &dto.GetFriendRequestsResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetFriends 获取好友列表
func (s *relationService) GetFriends(ctx context.Context, req *dto.GetFriendsRequest, userID uint) (*dto.GetFriendsResponse, error) {
	// 获取好友关系列表
	friends, total, err := s.friendRepo.GetFriends(userID, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	// 构建响应数据
	list := make([]dto.FriendItem, 0, len(friends))
	for _, friend := range friends {
		// 获取好友用户信息
		user, err := s.userRepo.FindByID(friend.TargetID)
		if err != nil {
			continue
		}

		// 添加到列表
		list = append(list, dto.FriendItem{
			ID:       user.ID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	return &dto.GetFriendsResponse{
		Total: int(total),
		List:  list,
	}, nil
}
