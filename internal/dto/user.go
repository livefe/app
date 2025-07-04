package dto

// UserBrief 用户简要信息
type UserBrief struct {
	ID       uint   `json:"id"`       // 用户ID
	Nickname string `json:"nickname"` // 用户昵称
	Avatar   string `json:"avatar"`   // 用户头像
	// 可以根据需要扩展更多字段
}

// VerificationType 验证码类型
type VerificationType string

// 验证码类型常量
const (
	VerificationTypeLogin      VerificationType = "login"      // 登录验证码
	VerificationTypeDeactivate VerificationType = "deactivate" // 注销账号验证码
)

// SendVerificationCodeRequest 发送验证码请求
type SendVerificationCodeRequest struct {
	Mobile string           `json:"mobile" binding:"required,mobile_cn"` // 手机号
	Type   VerificationType `json:"type" binding:"required"`             // 验证码类型
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

// DeactivateAccountRequest 注销账号请求
type DeactivateAccountRequest struct {
	UserID uint   `json:"user_id" binding:"required"`          // 用户ID
	Mobile string `json:"mobile" binding:"required,mobile_cn"` // 手机号
	Code   string `json:"code" binding:"required,len=6"`       // 验证码
}

// LogoutRequest 退出登录请求
type LogoutRequest struct {
	UserID uint   `json:"user_id" binding:"required"` // 用户ID
	Token  string `json:"-"`                          // JWT令牌，由处理器内部设置，不从请求中获取
}

// LogoutResponse 退出登录响应
type LogoutResponse struct {
	Message string `json:"message"` // 响应消息
}
