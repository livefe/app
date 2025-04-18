package dto

// SendVerificationCodeRequest 发送验证码请求
type SendVerificationCodeRequest struct {
	Mobile string `json:"mobile" binding:"required,mobile_cn"` // 手机号
}

// SendVerificationCodeResponse 发送验证码响应
type SendVerificationCodeResponse struct {
	Message string `json:"message"` // 响应消息
}

// VerificationCodeLoginRequest 验证码登录请求
type VerificationCodeLoginRequest struct {
	Mobile string `json:"mobile" binding:"required,mobile_cn"` // 手机号
	Code   string `json:"code" binding:"required,len=6"`       // 验证码
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"` // JWT令牌
	User  struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Mobile   string `json:"mobile"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	} `json:"user"` // 用户信息
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Mobile    string `json:"mobile"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}
