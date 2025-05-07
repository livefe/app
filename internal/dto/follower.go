package dto

import "time"

// 粉丝关注相关DTO

// FollowUserRequest 关注用户请求
type FollowUserRequest struct {
	TargetID uint `json:"target_id" binding:"required" validate:"required"`
}

// FollowUserResponse 关注用户响应
type FollowUserResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	TargetID  uint      `json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
}

// UnfollowUserRequest 取消关注用户请求
type UnfollowUserRequest struct {
	TargetID uint `json:"target_id" binding:"required" validate:"required"`
}

// GetFollowersRequest 获取粉丝列表请求
type GetFollowersRequest struct {
	UserID uint `json:"user_id" binding:"required" validate:"required"`
	Page   int  `json:"page" binding:"required" validate:"required,min=1"`
	Size   int  `json:"size" binding:"required" validate:"required,min=1,max=100"`
}

// GetFollowersResponse 获取粉丝列表响应
type GetFollowersResponse struct {
	Total int         `json:"total"`
	List  []UserBrief `json:"list"`
}

// GetFollowingRequest 获取关注列表请求
type GetFollowingRequest struct {
	UserID uint `json:"user_id" binding:"required" validate:"required"`
	Page   int  `json:"page" binding:"required" validate:"required,min=1"`
	Size   int  `json:"size" binding:"required" validate:"required,min=1,max=100"`
}

// GetFollowingResponse 获取关注列表响应
type GetFollowingResponse struct {
	Total int         `json:"total"`
	List  []UserBrief `json:"list"`
}
