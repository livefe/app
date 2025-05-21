package constant

import "time"

// 用户状态常量
const (
	// 用户状态：正常
	UserStatusNormal = 1
	// 用户状态：禁用
	UserStatusDisabled = 0
)

// 验证码相关常量
const (
	// 登录验证码Redis前缀
	VerificationCodePrefixLogin = "verification_code:login:"
	// 注销验证码Redis前缀
	VerificationCodePrefixDeactivate = "verification_code:deactivate:"
	// 验证码有效期（5分钟）
	VerificationCodeExpiration = 5 * time.Minute
	// 验证码长度
	VerificationCodeLength = 6
)

// 用户认证相关常量
const (
	// 令牌黑名单前缀
	TokenBlacklistPrefix = "token:blacklist:"
)

// 验证码类型
const (
	// 登录验证码类型
	VerificationTypeLogin = "login"
	// 注销验证码类型
	VerificationTypeDeactivate = "deactivate"
)

// 用户相关错误
var (
	// 用户不存在错误
	ErrUserNotFound = "用户不存在"
	// 验证码无效错误
	ErrInvalidCode = "验证码无效或已过期"
	// 注销失败错误
	ErrDeactivateFailed = "账号注销失败"
)
