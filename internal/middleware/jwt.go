package middleware

import (
	"errors"
	"net/http"
	"strings"

	"app/internal/constant"
	"app/pkg/jwt"
	"app/pkg/redis"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

// TokenBlacklistPrefix 令牌黑名单前缀
const TokenBlacklistPrefix = constant.TokenBlacklistPrefix

// AuthMiddleware 创建JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader(constant.AuthHeaderName)
		if authHeader == "" {
			response.Unauthorized(c, "未提供授权令牌", jwt.ErrTokenNotProvided)
			c.Abort()
			return
		}

		// 提取令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == constant.AuthHeaderPrefix) {
			response.Unauthorized(c, "无效的授权格式", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 检查令牌是否在黑名单中
		blacklistKey := TokenBlacklistPrefix + tokenString
		_, err := redis.Get(blacklistKey)
		if err == nil {
			// 令牌在黑名单中，拒绝访问
			response.Unauthorized(c, "令牌已失效，请重新登录", nil)
			c.Abort()
			return
		}

		// 解析令牌
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			var statusCode int
			var errorMsg string

			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				statusCode = http.StatusUnauthorized
				errorMsg = "令牌已过期"
			case errors.Is(err, jwt.ErrTokenInvalid):
				statusCode = http.StatusUnauthorized
				errorMsg = "无效的令牌"
			case errors.Is(err, jwt.ErrTokenNotProvided):
				statusCode = http.StatusUnauthorized
				errorMsg = "未提供授权令牌"
			default:
				statusCode = http.StatusInternalServerError
				errorMsg = "验证令牌时发生错误"
			}

			response.Fail(c, statusCode, errorMsg, err)
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		// 添加令牌ID到上下文，便于后续操作（如注销）
		if claims.ID != "" {
			c.Set("tokenID", claims.ID)
		}

		c.Next()
	}
}

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// GetUsernameFromContext 从上下文中获取用户名
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}
