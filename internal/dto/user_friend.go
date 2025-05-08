package dto

import "time"

// 好友关系相关DTO

// AddFriendRequest 添加好友请求
type AddFriendRequest struct {
	TargetID uint   `json:"target_id" binding:"required" validate:"required"`
	Message  string `json:"message" binding:"omitempty" validate:"omitempty,max=200"`
}

// AddFriendResponse 添加好友响应
type AddFriendResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	TargetID  uint      `json:"target_id"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// AcceptFriendRequest 接受好友请求
type AcceptFriendRequest struct {
	RequestID uint `json:"request_id" binding:"required" validate:"required"`
}

// RejectFriendRequest 拒绝好友请求
type RejectFriendRequest struct {
	RequestID uint `json:"request_id" binding:"required" validate:"required"`
}

// DeleteFriendRequest 删除好友请求
type DeleteFriendRequest struct {
	TargetID uint `json:"target_id" binding:"required" validate:"required"`
}

// GetFriendRequestsRequest 获取好友请求列表请求
type GetFriendRequestsRequest struct {
	Page int `json:"page" binding:"required" validate:"required,min=1"`
	Size int `json:"size" binding:"required" validate:"required,min=1,max=100"`
}

// FriendRequestItem 好友请求项
type FriendRequestItem struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// GetFriendRequestsResponse 获取好友请求列表响应
type GetFriendRequestsResponse struct {
	Total int                 `json:"total"`
	List  []FriendRequestItem `json:"list"`
}

// GetFriendsRequest 获取好友列表请求
type GetFriendsRequest struct {
	Page int `json:"page" binding:"required" validate:"required,min=1"`
	Size int `json:"size" binding:"required" validate:"required,min=1,max=100"`
}

// GetFriendsResponse 获取好友列表响应
type GetFriendsResponse struct {
	Total int          `json:"total"`
	List  []FriendItem `json:"list"`
}

// FriendItem 好友项
type FriendItem struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}