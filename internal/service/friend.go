package service

import (
	"app/internal/constant"
	"app/internal/dto"
	"app/internal/model"
	"app/internal/repository"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// FriendService 好友关系服务接口
type FriendService interface {
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

// friendService 好友关系服务实现
type friendService struct {
	friendRepo repository.FriendRepository
	userRepo   repository.UserRepository
}

// NewFriendService 创建好友关系服务实例
func NewFriendService(
	friendRepo repository.FriendRepository,
	userRepo repository.UserRepository,
) FriendService {
	return &friendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
	}
}

// AddFriend 添加好友
func (s *friendService) AddFriend(ctx context.Context, req *dto.AddFriendRequest, userID uint) (*dto.AddFriendResponse, error) {
	// 检查目标用户是否存在
	_, err := s.userRepo.FindByID(req.TargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("目标用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查是否已经是好友或已发送请求
	existingFriend, err := s.friendRepo.GetFriend(userID, req.TargetID)
	if err == nil {
		if existingFriend.Status == 1 {
			return nil, errors.New("已经是好友关系")
		}
		return nil, errors.New("已发送好友请求，等待对方确认")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询好友关系失败: %w", err)
	}

	// 创建好友请求
	friend := &model.Friend{
		UserID:   userID,
		TargetID: req.TargetID,
		Status:   0, // 待确认状态
	}

	err = s.friendRepo.CreateFriend(friend)
	if err != nil {
		return nil, fmt.Errorf("创建好友请求失败: %w", err)
	}

	return &dto.AddFriendResponse{
		ID:        friend.ID,
		UserID:    friend.UserID,
		TargetID:  friend.TargetID,
		Status:    friend.Status,
		CreatedAt: friend.CreatedAt,
	}, nil
}

// AcceptFriend 接受好友请求
func (s *friendService) AcceptFriend(ctx context.Context, req *dto.AcceptFriendRequest, userID uint) error {
	// 获取好友请求
	friend, err := s.friendRepo.GetFriendByID(req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友请求不存在")
		}
		return fmt.Errorf("查询好友请求失败: %w", err)
	}

	// 验证请求是否发给当前用户
	if friend.TargetID != userID {
		return errors.New("无权操作此好友请求")
	}

	// 验证请求状态
	if friend.Status != int(constant.FriendStatusPending) {
		return errors.New("该请求已处理")
	}

	// 更新好友状态为已接受
	err = s.friendRepo.UpdateFriendStatus(friend.ID, int(constant.FriendStatusConfirmed))
	if err != nil {
		return fmt.Errorf("接受好友请求失败: %w", err)
	}

	return nil
}

// RejectFriend 拒绝好友请求
func (s *friendService) RejectFriend(ctx context.Context, req *dto.RejectFriendRequest, userID uint) error {
	// 获取好友请求
	friend, err := s.friendRepo.GetFriendByID(req.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友请求不存在")
		}
		return fmt.Errorf("查询好友请求失败: %w", err)
	}

	// 验证请求是否发给当前用户
	if friend.TargetID != userID {
		return errors.New("无权操作此好友请求")
	}

	// 验证请求状态
	if friend.Status != 0 {
		return errors.New("该请求已处理")
	}

	// 删除好友请求
	err = s.friendRepo.DeleteFriend(friend.UserID, friend.TargetID)
	if err != nil {
		return fmt.Errorf("拒绝好友请求失败: %w", err)
	}

	return nil
}

// DeleteFriend 删除好友
func (s *friendService) DeleteFriend(ctx context.Context, req *dto.DeleteFriendRequest, userID uint) error {
	// 检查是否是好友关系
	friend, err := s.friendRepo.GetFriend(userID, req.TargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("不存在好友关系")
		}
		return fmt.Errorf("查询好友关系失败: %w", err)
	}

	// 验证好友状态
	if friend.Status != 1 {
		return errors.New("不是好友关系")
	}

	// 删除好友关系
	err = s.friendRepo.DeleteFriend(userID, req.TargetID)
	if err != nil {
		return fmt.Errorf("删除好友关系失败: %w", err)
	}

	return nil
}

// GetFriendRequests 获取好友请求列表
func (s *friendService) GetFriendRequests(ctx context.Context, req *dto.GetFriendRequestsRequest, userID uint) (*dto.GetFriendRequestsResponse, error) {
	// 获取好友请求列表
	requests, count, err := s.friendRepo.GetFriendRequests(userID, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("获取好友请求列表失败: %w", err)
	}

	// 构建好友请求项列表
	requestList := make([]dto.FriendRequestItem, 0, len(requests))
	for _, request := range requests {
		user, err := s.userRepo.FindByID(request.UserID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

		requestList = append(requestList, dto.FriendRequestItem{
			ID:        request.ID,
			UserID:    user.ID,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Message:   "", // 消息内容，可以扩展模型添加
			CreatedAt: request.CreatedAt,
		})
	}

	return &dto.GetFriendRequestsResponse{
		Total: int(count),
		List:  requestList,
	}, nil
}

// GetFriends 获取好友列表
func (s *friendService) GetFriends(ctx context.Context, req *dto.GetFriendsRequest, userID uint) (*dto.GetFriendsResponse, error) {
	// 获取好友关系列表
	friends, count, err := s.friendRepo.GetFriends(userID, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("获取好友列表失败: %w", err)
	}

	// 构建用户简要信息列表
	userList := make([]dto.UserBrief, 0, len(friends))
	for _, friend := range friends {
		// 确定好友ID（如果当前用户是UserID，那么好友是TargetID，反之亦然）
		friendID := friend.TargetID
		if friend.TargetID == userID {
			friendID = friend.UserID
		}

		user, err := s.userRepo.FindByID(friendID)
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

	return &dto.GetFriendsResponse{
		Total: int(count),
		List:  userList,
	}, nil
}
