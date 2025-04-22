package constant

// JWT相关常量

// TokenBlacklistPrefix 令牌黑名单前缀
const TokenBlacklistPrefix = "token:blacklist:"

// JWT错误类型
const (
	// ErrTokenExpired 令牌已过期错误
	ErrTokenExpired = "令牌已过期"
	// ErrTokenInvalid 无效的令牌错误
	ErrTokenInvalid = "无效的令牌"
	// ErrTokenNotProvided 未提供令牌错误
	ErrTokenNotProvided = "未提供令牌"
)

// JWT认证相关常量
const (
	// AuthHeaderName 认证头名称
	AuthHeaderName = "Authorization"
	// AuthHeaderPrefix 认证头前缀
	AuthHeaderPrefix = "Bearer"
)
