package dto

import "time"

// 社交动态相关DTO

// CreatePostRequest 创建动态请求
type CreatePostRequest struct {
	Content    string   `json:"content" validate:"required,max=1000"` // 动态内容
	ImageData  []string `json:"image_data"`                           // Base64编码的图片数据
	ImageIDs   []uint   `json:"image_ids"`                            // 已上传图片的ID列表
	Visibility int      `json:"visibility" validate:"min=0,max=2"`    // 可见性：0-公开，1-仅关注者可见，2-仅自己可见
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
