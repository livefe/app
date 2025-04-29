package dto

import "time"

// 社交关系相关DTO

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

// UserBrief 用户简要信息
type UserBrief struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// 位置分享相关DTO

// ShareLocationRequest 分享位置请求
type ShareLocationRequest struct {
	Latitude    float64    `json:"latitude" binding:"required" validate:"required"`
	Longitude   float64    `json:"longitude" binding:"required" validate:"required"`
	Address     string     `json:"address" binding:"required" validate:"required,max=255"`
	Description string     `json:"description" validate:"max=500"`
	Visibility  int        `json:"visibility" binding:"required" validate:"required,oneof=1 2 3"`
	ExpireTime  *time.Time `json:"expire_time"`
}

// ShareLocationResponse 分享位置响应
type ShareLocationResponse struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	Address     string     `json:"address"`
	Description string     `json:"description"`
	Visibility  int        `json:"visibility"`
	ExpireTime  *time.Time `json:"expire_time"`
	CreatedAt   time.Time  `json:"created_at"`
}

// GetNearbyUsersRequest 获取附近用户请求
type GetNearbyUsersRequest struct {
	Latitude  float64 `json:"latitude" binding:"required" validate:"required"`
	Longitude float64 `json:"longitude" binding:"required" validate:"required"`
	Radius    float64 `json:"radius" binding:"required" validate:"required,min=0.1,max=50"`
	Page      int     `json:"page" binding:"required" validate:"required,min=1"`
	Size      int     `json:"size" binding:"required" validate:"required,min=1,max=100"`
}

// GetNearbyUsersResponse 获取附近用户响应
type GetNearbyUsersResponse struct {
	Total int                `json:"total"`
	List  []NearbyUserDetail `json:"list"`
}

// NearbyUserDetail 附近用户详情
type NearbyUserDetail struct {
	UserBrief
	Distance    float64 `json:"distance"` // 距离，单位公里
	Address     string  `json:"address"`
	Description string  `json:"description"`
}

// 社交动态相关DTO

// CreatePostRequest 创建动态请求
type CreatePostRequest struct {
	Content    string `json:"content" binding:"required" validate:"required,max=2000"`
	Images     string `json:"images" validate:"max=1000"`
	LocationID *uint  `json:"location_id"`
	Visibility int    `json:"visibility" binding:"required" validate:"required,oneof=1 2 3"`
}

// CreatePostResponse 创建动态响应
type CreatePostResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content"`
	Images    string    `json:"images"`
	CreatedAt time.Time `json:"created_at"`
}

// GetPostsRequest 获取动态列表请求
type GetPostsRequest struct {
	UserID *uint `json:"user_id"` // 可选，为空表示获取关注用户的动态
	Page   int   `json:"page" binding:"required" validate:"required,min=1"`
	Size   int   `json:"size" binding:"required" validate:"required,min=1,max=100"`
}

// GetPostsResponse 获取动态列表响应
type GetPostsResponse struct {
	Total int          `json:"total"`
	List  []PostDetail `json:"list"`
}

// PostDetail 动态详情
type PostDetail struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	Content    string    `json:"content"`
	Images     string    `json:"images"`
	LocationID *uint     `json:"location_id"`
	Address    string    `json:"address,omitempty"`
	Likes      int       `json:"likes"`
	Comments   int       `json:"comments"`
	CreatedAt  time.Time `json:"created_at"`
}

// LikePostRequest 点赞动态请求
type LikePostRequest struct {
	PostID uint `json:"post_id" binding:"required" validate:"required"`
}

// CommentPostRequest 评论动态请求
type CommentPostRequest struct {
	PostID   uint   `json:"post_id" binding:"required" validate:"required"`
	Content  string `json:"content" binding:"required" validate:"required,max=500"`
	ParentID *uint  `json:"parent_id"` // 可选，回复某条评论
}

// CommentPostResponse 评论动态响应
type CommentPostResponse struct {
	ID        uint      `json:"id"`
	PostID    uint      `json:"post_id"`
	UserID    uint      `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Content   string    `json:"content"`
	ParentID  *uint     `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

// GetCommentsRequest 获取评论列表请求
type GetCommentsRequest struct {
	PostID uint `json:"post_id" binding:"required" validate:"required"`
	Page   int  `json:"page" binding:"required" validate:"required,min=1"`
	Size   int  `json:"size" binding:"required" validate:"required,min=1,max=100"`
}

// GetCommentsResponse 获取评论列表响应
type GetCommentsResponse struct {
	Total int             `json:"total"`
	List  []CommentDetail `json:"list"`
}

// CommentDetail 评论详情
type CommentDetail struct {
	ID        uint      `json:"id"`
	PostID    uint      `json:"post_id"`
	UserID    uint      `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Content   string    `json:"content"`
	ParentID  *uint     `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}
