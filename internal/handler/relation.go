package handler

import (
	"app/internal/dto"
	"app/internal/service"
	"app/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RelationHandler 用户关系处理器
type RelationHandler struct {
	relationService service.RelationService
}

// NewRelationHandler 创建用户关系处理器实例
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

	res, err := h.relationService.GetFollowing(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "获取关注列表失败", err)
		return
	}

	response.Success(c, "获取关注列表成功", res)
}

// AddFriend 添加好友
func (h *RelationHandler) AddFriend(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	res, err := h.relationService.AddFriend(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "添加好友失败", err)
		return
	}

	response.Success(c, "好友请求已发送", res)
}

// AcceptFriend 接受好友请求
func (h *RelationHandler) AcceptFriend(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.AcceptFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	err := h.relationService.AcceptFriend(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "接受好友请求失败", err)
		return
	}

	response.Success(c, "已接受好友请求", nil)
}

// RejectFriend 拒绝好友请求
func (h *RelationHandler) RejectFriend(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.RejectFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	err := h.relationService.RejectFriend(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "拒绝好友请求失败", err)
		return
	}

	response.Success(c, "已拒绝好友请求", nil)
}

// DeleteFriend 删除好友
func (h *RelationHandler) DeleteFriend(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	var req dto.DeleteFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误", err)
		return
	}

	err := h.relationService.DeleteFriend(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "删除好友失败", err)
		return
	}

	response.Success(c, "已删除好友", nil)
}

// GetFriendRequests 获取好友请求列表
func (h *RelationHandler) GetFriendRequests(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	req := &dto.GetFriendRequestsRequest{
		Page: page,
		Size: size,
	}

	res, err := h.relationService.GetFriendRequests(c.Request.Context(), req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "获取好友请求列表失败", err)
		return
	}

	response.Success(c, "获取好友请求列表成功", res)
}

// GetFriends 获取好友列表
func (h *RelationHandler) GetFriends(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录", nil)
		return
	}

	// 解析请求参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	req := &dto.GetFriendsRequest{
		Page: page,
		Size: size,
	}

	res, err := h.relationService.GetFriends(c.Request.Context(), req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "获取好友列表失败", err)
		return
	}

	response.Success(c, "获取好友列表成功", res)
}
