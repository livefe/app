package dto

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Nickname   string `json:"nickname" validate:"max=50"` // 用户昵称
	AvatarData string `json:"avatar_data"`                // Base64编码的头像图片数据
}

// UserProfileResponse 用户资料响应
type UserProfileResponse struct {
	ID       uint   `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 用户昵称
	Mobile   string `json:"mobile"`   // 手机号
	Avatar   string `json:"avatar"`   // 头像URL
}
