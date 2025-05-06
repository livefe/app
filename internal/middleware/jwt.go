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

// AuthMiddleware 创建JWT认证中间件，验证请求中的令牌并提取用户信息
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(constant.AuthHeaderName)
		if authHeader == "" {
			response.Unauthorized(c, "未提供授权令牌", jwt.ErrTokenNotProvided)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == constant.AuthHeaderPrefix) {
			response.Unauthorized(c, "无效的授权格式", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		blacklistKey := constant.TokenBlacklistPrefix + tokenString
		_, err := redis.Get(blacklistKey)
		if err == nil {
			response.Unauthorized(c, "令牌已失效，请重新登录", nil)
			c.Abort()
			return
		}

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

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		if claims.ID != "" {
			c.Set("tokenID", claims.ID)
		}

		c.Next()
	}
}
