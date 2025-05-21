package constant

import "time"

// 用户相关常量

// 用户状态常量
const (
	// UserStatusNormal 用户状态：正常
	UserStatusNormal = 1
	// UserStatusDisabled 用户状态：禁用
	UserStatusDisabled = 0
)

// 验证码相关常量
const (
	// VerificationCodePrefixLogin 登录验证码Redis前缀
	VerificationCodePrefixLogin = "verification_code:login:"
	// VerificationCodePrefixDeactivate 注销验证码Redis前缀
	VerificationCodePrefixDeactivate = "verification_code:deactivate:"
	// VerificationCodeExpiration 验证码有效期（5分钟）
	VerificationCodeExpiration = 5 * time.Minute
	// VerificationCodeLength 验证码长度
	VerificationCodeLength = 6
)

// 用户认证相关常量
const (
	// TokenBlacklistPrefix 令牌黑名单前缀
	TokenBlacklistPrefix = "token:blacklist:"
)

// 验证码类型
const (
	// VerificationTypeLogin 登录验证码类型
	VerificationTypeLogin = "login"
	// VerificationTypeDeactivate 注销验证码类型
	VerificationTypeDeactivate = "deactivate"
)

// 用户相关错误
var (
	// ErrUserNotFound 用户不存在错误
	ErrUserNotFound = "用户不存在"
	// ErrInvalidCode 验证码无效错误
	ErrInvalidCode = "验证码无效或已过期"
	// ErrDeactivateFailed 注销失败错误
	ErrDeactivateFailed = "账号注销失败"
)
