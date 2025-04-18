package middleware

import (
	"net/http"
	"strings"

	"app/pkg/jwt"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 创建JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未提供授权令牌", jwt.ErrTokenNotProvided)
			c.Abort()
			return
		}

		// 提取令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "无效的授权格式", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析令牌
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			var statusCode int
			var errorMsg string

			switch err {
			case jwt.ErrTokenExpired:
				statusCode = http.StatusUnauthorized
				errorMsg = "令牌已过期"
			case jwt.ErrTokenInvalid:
				statusCode = http.StatusUnauthorized
				errorMsg = "无效的令牌"
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
