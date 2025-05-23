package handler

import (
	"app/internal/dto"
	"app/internal/service"
	"app/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PostHandler 动态处理器
type PostHandler struct {
	postService service.PostService
}

// NewPostHandler 创建动态处理器实例
func NewPostHandler(postService service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// CreatePost 创建动态
func (h *PostHandler) CreatePost(c *gin.Context) {
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

	// 调用服务创建动态
	res, err := h.postService.CreatePost(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "创建动态失败", err)
		return
	}

	response.Success(c, "创建动态成功", res)
}

// GetPosts 获取动态列表
func (h *PostHandler) GetPosts(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	// 解析用户ID参数（可选）
	var targetUserID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := strconv.ParseUint(userIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			targetUserID = &uid
		}
	}

	req := &dto.GetPostsRequest{
		UserID: targetUserID,
		Page:   page,
		Size:   size,
	}

	// 调用服务
	res, err := h.postService.GetPosts(c.Request.Context(), req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "获取动态列表失败", err)
		return
	}

	response.Success(c, "获取动态列表成功", res)
}

// LikePost 点赞动态
func (h *PostHandler) LikePost(c *gin.Context) {
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
	err := h.postService.LikePost(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "点赞失败", err)
		return
	}

	response.Success(c, "点赞成功", nil)
}

// CommentPost 评论动态
func (h *PostHandler) CommentPost(c *gin.Context) {
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
	res, err := h.postService.CommentPost(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "评论失败", err)
		return
	}

	response.Success(c, "评论成功", res)
}

// GetComments 获取评论列表
func (h *PostHandler) GetComments(c *gin.Context) {
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
	res, err := h.postService.GetComments(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "获取评论列表失败", err)
		return
	}

	response.Success(c, "获取评论列表成功", res)
}
