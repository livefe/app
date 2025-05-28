package service

import (
	"app/internal/dto"
	"app/internal/model"
	"app/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// PostService 动态服务接口
type PostService interface {
	// CreatePost 创建动态
	CreatePost(ctx context.Context, req *dto.CreatePostRequest, userID uint) (*dto.CreatePostResponse, error)
	// GetPosts 获取动态列表
	GetPosts(ctx context.Context, req *dto.GetPostsRequest, userID uint) (*dto.GetPostsResponse, error)
	// LikePost 点赞动态
	LikePost(ctx context.Context, req *dto.LikePostRequest, userID uint) error
	// CommentPost 评论动态
	CommentPost(ctx context.Context, req *dto.CommentPostRequest, userID uint) (*dto.CommentPostResponse, error)
	// GetComments 获取评论列表
	GetComments(ctx context.Context, req *dto.GetCommentsRequest) (*dto.GetCommentsResponse, error)
}

// postService 动态服务实现
type postService struct {
	postRepo      repository.PostRepository
	commentRepo   repository.PostCommentRepository
	userRepo      repository.UserRepository
	postImageRepo repository.PostImageRepository
	imageService  ImageService
}

// NewPostService 创建动态服务实例
func NewPostService(
	postRepo repository.PostRepository,
	commentRepo repository.PostCommentRepository,
	userRepo repository.UserRepository,
	postImageRepo repository.PostImageRepository,
	imageService ImageService,
) PostService {
	return &postService{
		postRepo:      postRepo,
		commentRepo:   commentRepo,
		userRepo:      userRepo,
		postImageRepo: postImageRepo,
		imageService:  imageService,
	}
}

// CreatePost 创建动态
func (s *postService) CreatePost(ctx context.Context, req *dto.CreatePostRequest, userID uint) (*dto.CreatePostResponse, error) {
	// 创建动态
	post := &model.Post{
		UserID:     userID,
		Content:    req.Content,
		Visibility: req.Visibility, // 使用dto中的可见性值，对应constant.Visibility类型
		Likes:      0,
		Comments:   0,
	}

	// 保存动态基本信息
	err := s.postRepo.CreatePost(post)
	if err != nil {
		return nil, fmt.Errorf("创建动态失败: %w", err)
	}

	// 处理图片上传
	var imageURLs []string

	// 处理已上传的图片ID列表
	if len(req.ImageIDs) > 0 {
		for _, imageID := range req.ImageIDs {
			// 移动图片到动态并关联
			postImage, err := s.imageService.MoveImageToPost(ctx, imageID, post.ID, userID)
			if err != nil {
				fmt.Printf("关联图片失败: %v\n", err)
				continue // 跳过关联失败的图片
			}

			// 添加图片URL到列表
			imageURLs = append(imageURLs, postImage.URL)
		}
	}

	return &dto.CreatePostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	}, nil
}

// GetPosts 获取动态列表
func (s *postService) GetPosts(ctx context.Context, req *dto.GetPostsRequest, userID uint) (*dto.GetPostsResponse, error) {
	var posts []model.Post
	var count int64
	var err error

	// 根据请求参数获取不同的动态列表
	if req.UserID != nil && *req.UserID > 0 {
		// 获取指定用户的动态，传递当前用户ID作为查看者ID
		posts, count, err = s.postRepo.GetUserPosts(*req.UserID, req.Page, req.Size, userID)
	} else {
		// 获取关注用户的动态
		posts, count, err = s.postRepo.GetFollowingPosts(userID, req.Page, req.Size)
	}

	if err != nil {
		return nil, fmt.Errorf("获取动态列表失败: %w", err)
	}

	// 构建动态信息列表
	postList := make([]dto.PostDetail, 0, len(posts))
	for _, post := range posts {
		user, err := s.userRepo.FindByID(post.UserID)
		if err != nil {
			continue // 跳过获取失败的用户
		}

		// 获取动态图片
		var images string
		// 从图片关联中获取
		postImages, err := s.postImageRepo.GetPostImages(post.ID)
		if err == nil && len(postImages) > 0 {
			imageURLs := make([]string, len(postImages))
			for i, img := range postImages {
				imageURLs[i] = img.URL
			}
			images = strings.Join(imageURLs, ",")
		}

		postList = append(postList, dto.PostDetail{
			ID:        post.ID,
			UserID:    post.UserID,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Content:   post.Content,
			Images:    images,
			Likes:     post.Likes,
			Comments:  post.Comments,
			CreatedAt: post.CreatedAt,
		})
	}

	return &dto.GetPostsResponse{
		Total: int(count),
		List:  postList,
	}, nil
}

// LikePost 点赞动态
func (s *postService) LikePost(ctx context.Context, req *dto.LikePostRequest, userID uint) error {
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

// CommentPost 评论动态
func (s *postService) CommentPost(ctx context.Context, req *dto.CommentPostRequest, userID uint) (*dto.CommentPostResponse, error) {
	// 检查动态是否存在
	_, err := s.postRepo.GetPost(req.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("动态不存在")
		}
		return nil, fmt.Errorf("查询动态失败: %w", err)
	}

	// 创建评论
	comment := &model.PostComment{
		PostID:   req.PostID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	// 使用事务创建评论
	err = s.commentRepo.CreateCommentWithTransaction(comment, req.PostID)
	if err != nil {
		return nil, err
	}

	// 获取用户信息以返回昵称和头像
	user, _ := s.userRepo.FindByID(userID)

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
func (s *postService) GetComments(ctx context.Context, req *dto.GetCommentsRequest) (*dto.GetCommentsResponse, error) {
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
