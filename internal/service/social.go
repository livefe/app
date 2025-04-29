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

// SocialService 社交服务接口
type SocialService interface {
	// FollowUser 关注用户
	FollowUser(ctx context.Context, req *dto.FollowUserRequest, userID uint) (*dto.FollowUserResponse, error)
	// UnfollowUser 取消关注用户
	UnfollowUser(ctx context.Context, req *dto.UnfollowUserRequest, userID uint) error
	// GetFollowers 获取粉丝列表
	GetFollowers(ctx context.Context, req *dto.GetFollowersRequest) (*dto.GetFollowersResponse, error)
	// GetFollowing 获取关注列表
	GetFollowing(ctx context.Context, req *dto.GetFollowingRequest) (*dto.GetFollowingResponse, error)
	// ShareLocation 分享位置
	ShareLocation(ctx context.Context, req *dto.ShareLocationRequest, userID uint) (*dto.ShareLocationResponse, error)
	// GetNearbyUsers 获取附近用户
	GetNearbyUsers(ctx context.Context, req *dto.GetNearbyUsersRequest, userID uint) (*dto.GetNearbyUsersResponse, error)
	// CreatePost 创建社交动态
	CreatePost(ctx context.Context, req *dto.CreatePostRequest, userID uint) (*dto.CreatePostResponse, error)
	// GetPosts 获取社交动态列表
	GetPosts(ctx context.Context, req *dto.GetPostsRequest, userID uint) (*dto.GetPostsResponse, error)
	// LikePost 点赞社交动态
	LikePost(ctx context.Context, req *dto.LikePostRequest, userID uint) error
	// CommentPost 评论社交动态
	CommentPost(ctx context.Context, req *dto.CommentPostRequest, userID uint) (*dto.CommentPostResponse, error)
	// GetComments 获取评论列表
	GetComments(ctx context.Context, req *dto.GetCommentsRequest) (*dto.GetCommentsResponse, error)
}

// socialService 社交服务实现
type socialService struct {
	relationRepo      repository.SocialRelationRepository
	locationShareRepo repository.SocialLocationShareRepository
	postRepo          repository.SocialPostRepository
	commentRepo       repository.SocialCommentRepository
	userRepo          repository.UserRepository
}

// NewSocialService 创建社交服务实例
func NewSocialService(
	relationRepo repository.SocialRelationRepository,
	locationShareRepo repository.SocialLocationShareRepository,
	postRepo repository.SocialPostRepository,
	commentRepo repository.SocialCommentRepository,
	userRepo repository.UserRepository,
) SocialService {
	return &socialService{
		relationRepo:      relationRepo,
		locationShareRepo: locationShareRepo,
		postRepo:          postRepo,
		commentRepo:       commentRepo,
		userRepo:          userRepo,
	}
}

// FollowUser 关注用户
func (s *socialService) FollowUser(ctx context.Context, req *dto.FollowUserRequest, userID uint) (*dto.FollowUserResponse, error) {
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
	relation := &model.SocialRelation{
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
func (s *socialService) UnfollowUser(ctx context.Context, req *dto.UnfollowUserRequest, userID uint) error {
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
func (s *socialService) GetFollowers(ctx context.Context, req *dto.GetFollowersRequest) (*dto.GetFollowersResponse, error) {
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
func (s *socialService) GetFollowing(ctx context.Context, req *dto.GetFollowingRequest) (*dto.GetFollowingResponse, error) {
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

// ShareLocation 分享位置
func (s *socialService) ShareLocation(ctx context.Context, req *dto.ShareLocationRequest, userID uint) (*dto.ShareLocationResponse, error) {
	// 创建位置分享记录
	location := &model.SocialLocationShare{
		UserID:      userID,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Description: req.Description,
		Visibility:  req.Visibility,
		ExpireTime:  req.ExpireTime,
	}

	err := s.locationShareRepo.CreateLocationShare(location)
	if err != nil {
		return nil, fmt.Errorf("创建位置分享失败: %w", err)
	}

	return &dto.ShareLocationResponse{
		ID:          location.ID,
		UserID:      location.UserID,
		Latitude:    location.Latitude,
		Longitude:   location.Longitude,
		Address:     location.Address,
		Description: location.Description,
		Visibility:  location.Visibility,
		ExpireTime:  location.ExpireTime,
		CreatedAt:   location.CreatedAt,
	}, nil
}

// GetNearbyUsers 获取附近用户
func (s *socialService) GetNearbyUsers(ctx context.Context, req *dto.GetNearbyUsersRequest, userID uint) (*dto.GetNearbyUsersResponse, error) {
	// 获取附近的位置分享
	locations, count, err := s.locationShareRepo.GetNearbyUsers(userID, req.Latitude, req.Longitude, req.Radius, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("获取附近用户失败: %w", err)
	}

	// 构建用户位置信息列表
	userList := make([]dto.NearbyUserDetail, 0, len(locations))
	for _, location := range locations {
		user, err := s.userRepo.FindByID(location.UserID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

		userList = append(userList, dto.NearbyUserDetail{
			UserBrief: dto.UserBrief{
				ID:       user.ID,
				Nickname: user.Nickname,
				Avatar:   user.Avatar,
			},
			Distance:    0, // 距离信息需要从查询结果中获取，这里简化处理
			Address:     location.Address,
			Description: location.Description,
		})
	}

	return &dto.GetNearbyUsersResponse{
		Total: int(count),
		List:  userList,
	}, nil
}

// CreatePost 创建社交动态
func (s *socialService) CreatePost(ctx context.Context, req *dto.CreatePostRequest, userID uint) (*dto.CreatePostResponse, error) {
	// 创建社交动态
	post := &model.SocialPost{
		UserID:     userID,
		Content:    req.Content,
		Images:     req.Images,
		Visibility: req.Visibility,
		LocationID: req.LocationID,
		Likes:      0,
		Comments:   0,
	}

	err := s.postRepo.CreatePost(post)
	if err != nil {
		return nil, fmt.Errorf("创建社交动态失败: %w", err)
	}

	return &dto.CreatePostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		Content:   post.Content,
		Images:    post.Images,
		CreatedAt: post.CreatedAt,
	}, nil
}

// GetPosts 获取社交动态列表
func (s *socialService) GetPosts(ctx context.Context, req *dto.GetPostsRequest, userID uint) (*dto.GetPostsResponse, error) {
	var posts []model.SocialPost
	var count int64
	var err error

	// 根据请求参数获取不同的动态列表
	if req.UserID != nil && *req.UserID > 0 {
		// 获取指定用户的动态
		posts, count, err = s.postRepo.GetUserPosts(*req.UserID, req.Page, req.Size)
	} else {
		// 获取关注用户的动态
		posts, count, err = s.postRepo.GetFollowingPosts(userID, req.Page, req.Size)
	}

	if err != nil {
		return nil, fmt.Errorf("获取社交动态列表失败: %w", err)
	}

	// 构建动态信息列表
	postList := make([]dto.PostDetail, 0, len(posts))
	for _, post := range posts {
		user, err := s.userRepo.FindByID(post.UserID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

		postList = append(postList, dto.PostDetail{
			ID:         post.ID,
			UserID:     post.UserID,
			Nickname:   user.Nickname,
			Avatar:     user.Avatar,
			Content:    post.Content,
			Images:     post.Images,
			LocationID: post.LocationID,
			Likes:      post.Likes,
			Comments:   post.Comments,
			CreatedAt:  post.CreatedAt,
		})

		// 如果有位置信息，可以在这里获取地址信息
		if post.LocationID != nil && *post.LocationID > 0 {
			// 这里可以通过locationShareRepo获取位置信息并填充Address字段
			// 简化处理，暂不实现
		}
	}

	return &dto.GetPostsResponse{
		Total: int(count),
		List:  postList,
	}, nil
}

// LikePost 点赞社交动态
func (s *socialService) LikePost(ctx context.Context, req *dto.LikePostRequest, userID uint) error {
	// 检查动态是否存在
	_, err := s.postRepo.GetPost(req.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("动态不存在")
		}
		return fmt.Errorf("查询动态失败: %w", err)
	}

	// 增加点赞数
	err = s.postRepo.IncrementPostLikes(req.PostID)
	if err != nil {
		return fmt.Errorf("点赞失败: %w", err)
	}

	return nil
}

// CommentPost 评论社交动态
func (s *socialService) CommentPost(ctx context.Context, req *dto.CommentPostRequest, userID uint) (*dto.CommentPostResponse, error) {
	// 检查动态是否存在
	_, err := s.postRepo.GetPost(req.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("动态不存在")
		}
		return nil, fmt.Errorf("查询动态失败: %w", err)
	}

	// 创建评论
	comment := &model.SocialComment{
		PostID:   req.PostID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	err = s.commentRepo.CreateComment(comment)
	if err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	// 增加评论数
	err = s.postRepo.IncrementPostComments(req.PostID)
	if err != nil {
		// 评论已创建，但增加评论数失败，记录错误但不影响返回
		// 实际项目中应该使用事务或消息队列确保一致性
		fmt.Printf("增加评论数失败: %v\n", err)
	}

	// 获取用户信息以返回昵称和头像
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		// 用户信息获取失败，但不影响评论创建，使用默认值
		fmt.Printf("获取用户信息失败: %v\n", err)
	}

	var nickname, avatar string
	if user != nil {
		nickname = user.Nickname
		avatar = user.Avatar
	}

	return &dto.CommentPostResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		Nickname:  nickname,
		Avatar:    avatar,
		Content:   comment.Content,
		ParentID:  comment.ParentID,
		CreatedAt: comment.CreatedAt,
	}, nil
}

// GetComments 获取评论列表
func (s *socialService) GetComments(ctx context.Context, req *dto.GetCommentsRequest) (*dto.GetCommentsResponse, error) {
	// 获取评论列表
	comments, count, err := s.commentRepo.GetPostComments(req.PostID, req.Page, req.Size)
	if err != nil {
		return nil, fmt.Errorf("获取评论列表失败: %w", err)
	}

	// 构建评论信息列表
	commentList := make([]dto.CommentDetail, 0, len(comments))
	for _, comment := range comments {
		user, err := s.userRepo.FindByID(comment.UserID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

		commentList = append(commentList, dto.CommentDetail{
			ID:        comment.ID,
			PostID:    comment.PostID,
			UserID:    comment.UserID,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Content:   comment.Content,
			ParentID:  comment.ParentID,
			CreatedAt: comment.CreatedAt,
		})
	}

	return &dto.GetCommentsResponse{
		Total: int(count),
		List:  commentList,
	}, nil
}
