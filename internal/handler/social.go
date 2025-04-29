package handler

import (
	"app/internal/dto"
	"app/internal/service"
	"app/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SocialHandler 社交处理器
type SocialHandler struct {
	socialService service.SocialService
}

// NewSocialHandler 创建社交处理器实例
func NewSocialHandler(socialService service.SocialService) *SocialHandler {
	return &SocialHandler{
		socialService: socialService,
	}
}

// FollowUser 关注用户
func (h *SocialHandler) FollowUser(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.FollowUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	// 调用服务
	res, err := h.socialService.FollowUser(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "关注用户失败", err)
		return
	}

	response.Success(c, "关注成功", res)
}

// UnfollowUser 取消关注用户
func (h *SocialHandler) UnfollowUser(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.UnfollowUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	// 调用服务
	err := h.socialService.UnfollowUser(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "取消关注失败", err)
		return
	}

	response.Success(c, "取消关注成功", nil)
}

// GetFollowers 获取粉丝列表
func (h *SocialHandler) GetFollowers(c *gin.Context) {
	// 解析请求参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误", err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	req := &dto.GetFollowersRequest{
		UserID: uint(userID),
		Page:   page,
		Size:   size,
	}

	// 调用服务
	res, err := h.socialService.GetFollowers(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "获取粉丝列表失败", err)
		return
	}

	response.Success(c, "获取粉丝列表成功", res)
}

// GetFollowing 获取关注列表
func (h *SocialHandler) GetFollowing(c *gin.Context) {
	// 解析请求参数
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误", err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	req := &dto.GetFollowingRequest{
		UserID: uint(userID),
		Page:   page,
		Size:   size,
	}

	// 调用服务
	res, err := h.socialService.GetFollowing(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "获取关注列表失败", err)
		return
	}

	response.Success(c, "获取关注列表成功", res)
}

// ShareLocation 分享位置
func (h *SocialHandler) ShareLocation(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.ShareLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	// 调用服务
	res, err := h.socialService.ShareLocation(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "分享位置失败", err)
		return
	}

	response.Success(c, "分享位置成功", res)
}

// GetNearbyUsers 获取附近用户
func (h *SocialHandler) GetNearbyUsers(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.GetNearbyUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	// 调用服务
	res, err := h.socialService.GetNearbyUsers(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "获取附近用户失败", err)
		return
	}

	response.Success(c, "获取附近用户成功", res)
}

// CreatePost 创建社交动态
func (h *SocialHandler) CreatePost(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	// 调用服务
	res, err := h.socialService.CreatePost(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "创建动态失败", err)
		return
	}

	response.Success(c, "创建动态成功", res)
}

// GetPosts 获取社交动态列表
func (h *SocialHandler) GetPosts(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	userIDStr := c.Query("user_id")
	var userIDPtr *uint
	if userIDStr != "" {
		userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "用户ID格式错误", err)
			return
		}
		userIDUint32 := uint(userIDUint)
		userIDPtr = &userIDUint32
	}

	req := &dto.GetPostsRequest{
		UserID: userIDPtr,
		Page:   page,
		Size:   size,
	}

	// 调用服务
	res, err := h.socialService.GetPosts(c.Request.Context(), req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "获取动态列表失败", err)
		return
	}

	response.Success(c, "获取动态列表成功", res)
}

// LikePost 点赞社交动态
func (h *SocialHandler) LikePost(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.LikePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	// 调用服务
	err := h.socialService.LikePost(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "点赞失败", err)
		return
	}

	response.Success(c, "点赞成功", nil)
}

// CommentPost 评论社交动态
func (h *SocialHandler) CommentPost(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.CommentPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	// 调用服务
	res, err := h.socialService.CommentPost(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "评论失败", err)
		return
	}

	response.Success(c, "评论成功", res)
}

// GetComments 获取评论列表
func (h *SocialHandler) GetComments(c *gin.Context) {
	// 解析请求参数
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "动态ID格式错误", err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	req := &dto.GetCommentsRequest{
		PostID: uint(postID),
		Page:   page,
		Size:   size,
	}

	// 调用服务
	res, err := h.socialService.GetComments(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "获取评论列表失败", err)
		return
	}

	response.Success(c, "获取评论列表成功", res)
}
