// Package jwt 提供JWT令牌的生成、解析和验证功能
package jwt

import (
	"errors"
	"fmt"
	"time"

	"app/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT错误定义
var (
	ErrTokenExpired     = errors.New("令牌已过期") // 令牌已过期
	ErrTokenInvalid     = errors.New("无效的令牌") // 令牌无效
	ErrTokenNotProvided = errors.New("未提供令牌") // 未提供令牌
)

// JWT认证相关常量
const (
	// TokenBlacklistPrefix 令牌黑名单前缀
	TokenBlacklistPrefix = "token:blacklist:"
	// AuthHeaderName 认证头名称
	AuthHeaderName = "Authorization"
	// AuthHeaderPrefix 认证头前缀
	AuthHeaderPrefix = "Bearer"
)

// CustomClaims 自定义JWT声明结构体
type CustomClaims struct {
	UserID               uint   `json:"user_id"`  // 用户ID
	Username             string `json:"username"` // 用户名
	jwt.RegisteredClaims        // 标准JWT声明
}

// GenerateToken 生成包含用户信息的JWT令牌
func GenerateToken(userID uint, username string, _ string) (string, error) {
	jwtConfig := config.GetJWTConfig()

	expDuration, err := time.ParseDuration(jwtConfig.ExpiresTime)
	if err != nil {
		return "", fmt.Errorf("解析过期时间失败: %w", err)
	}

	now := time.Now()
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jwtConfig.Issuer,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtConfig.SecretKey))
	if err != nil {
		return "", fmt.Errorf("签名令牌失败: %w", err)
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌并提取其中的声明信息
func ParseToken(tokenString string) (*CustomClaims, error) {
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	jwtConfig := config.GetJWTConfig()
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ValidateToken 验证JWT令牌有效性
func ValidateToken(tokenString string) (bool, error) {
	_, err := ParseToken(tokenString)
	return err == nil, err
}

// RefreshToken 刷新JWT令牌
func RefreshToken(tokenString string) (string, error) {
	claims, err := parseTokenWithoutValidation(tokenString)
	if err != nil {
		return "", fmt.Errorf("无法解析原令牌: %w", err)
	}

	// 检查令牌是否已过期太久（超过7天不允许刷新）
	if claims.ExpiresAt != nil {
		expTime := claims.ExpiresAt.Time
		if time.Since(expTime) > 7*24*time.Hour {
			return "", fmt.Errorf("令牌已过期太久，无法刷新")
		}
	}

	return GenerateToken(claims.UserID, claims.Username, "")
}

// parseTokenWithoutValidation 解析JWT令牌但不验证过期时间
func parseTokenWithoutValidation(tokenString string) (*CustomClaims, error) {
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	jwtConfig := config.GetJWTConfig()
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(jwtConfig.SecretKey), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenInvalid
		}
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
