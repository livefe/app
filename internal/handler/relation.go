package handler

import (
	"app/internal/dto"
	"app/internal/service"
	"app/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RelationHandler 关系处理器
type RelationHandler struct {
	relationService service.RelationService
}

// NewRelationHandler 创建关系处理器实例
func NewRelationHandler(relationService service.RelationService) *RelationHandler {
	return &RelationHandler{
		relationService: relationService,
	}
}

// FollowUser 关注用户
func (h *RelationHandler) FollowUser(c *gin.Context) {
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
	res, err := h.relationService.FollowUser(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "关注用户失败", err)
		return
	}

	response.Success(c, "关注成功", res)
}

// UnfollowUser 取消关注用户
func (h *RelationHandler) UnfollowUser(c *gin.Context) {
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
	err := h.relationService.UnfollowUser(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "取消关注失败", err)
		return
	}

	response.Success(c, "取消关注成功", nil)
}

// GetFollowers 获取粉丝列表
func (h *RelationHandler) GetFollowers(c *gin.Context) {
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
	res, err := h.relationService.GetFollowers(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "获取粉丝列表失败", err)
		return
	}

	response.Success(c, "获取粉丝列表成功", res)
}

// GetFollowing 获取关注列表
func (h *RelationHandler) GetFollowing(c *gin.Context) {
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
	res, err := h.relationService.GetFollowing(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "获取关注列表失败", err)
		return
	}

	response.Success(c, "获取关注列表成功", res)
}
