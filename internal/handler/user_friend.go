package handler

import (
	"app/internal/dto"
	"app/internal/service"
	"app/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserFriendHandler 好友关系处理器
type UserFriendHandler struct {
	friendService service.UserFriendService
}

// NewUserFriendHandler 创建好友关系处理器实例
func NewUserFriendHandler(friendService service.UserFriendService) *UserFriendHandler {
	return &UserFriendHandler{
		friendService: friendService,
	}
}

// AddFriend 添加好友
func (h *UserFriendHandler) AddFriend(c *gin.Context) {
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

	// 调用服务
	res, err := h.friendService.AddFriend(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "添加好友失败", err)
		return
	}

	response.Success(c, "好友请求已发送", res)
}

// AcceptFriend 接受好友请求
func (h *UserFriendHandler) AcceptFriend(c *gin.Context) {
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

	// 调用服务
	err := h.friendService.AcceptFriend(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "接受好友请求失败", err)
		return
	}

	response.Success(c, "已接受好友请求", nil)
}

// RejectFriend 拒绝好友请求
func (h *UserFriendHandler) RejectFriend(c *gin.Context) {
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

	// 调用服务
	err := h.friendService.RejectFriend(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "拒绝好友请求失败", err)
		return
	}

	response.Success(c, "已拒绝好友请求", nil)
}

// DeleteFriend 删除好友
func (h *UserFriendHandler) DeleteFriend(c *gin.Context) {
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

	// 调用服务
	err := h.friendService.DeleteFriend(c.Request.Context(), &req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "删除好友失败", err)
		return
	}

	response.Success(c, "已删除好友关系", nil)
}

// GetFriendRequests 获取好友请求列表
func (h *UserFriendHandler) GetFriendRequests(c *gin.Context) {
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

	// 调用服务
	res, err := h.friendService.GetFriendRequests(c.Request.Context(), req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "获取好友请求列表失败", err)
		return
	}

	response.Success(c, "获取好友请求列表成功", res)
}

// GetFriends 获取好友列表
func (h *UserFriendHandler) GetFriends(c *gin.Context) {
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

	// 调用服务
	res, err := h.friendService.GetFriends(c.Request.Context(), req, userID.(uint))
	if err != nil {
		response.InternalServerError(c, "获取好友列表失败", err)
		return
	}

	response.Success(c, "获取好友列表成功", res)
}