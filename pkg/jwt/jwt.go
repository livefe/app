// Package jwt 提供JWT令牌的生成、解析和验证功能
package jwt

import (
	"errors"
	"fmt"
	"time"

	"app/config"
	"app/internal/constant"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 定义JWT相关错误类型，便于统一错误处理
var (
	// ErrTokenExpired 表示令牌已过期
	ErrTokenExpired = errors.New(constant.ErrTokenExpired)
	// ErrTokenInvalid 表示令牌无效
	ErrTokenInvalid = errors.New(constant.ErrTokenInvalid)
	// ErrTokenNotProvided 表示未提供令牌
	ErrTokenNotProvided = errors.New(constant.ErrTokenNotProvided)
)

// CustomClaims 自定义JWT声明结构体
// 包含用户信息和标准JWT声明
// 用于在令牌中存储和传递用户相关信息
type CustomClaims struct {
	UserID               uint   `json:"user_id"`  // 用户ID
	Username             string `json:"username"` // 用户名
	jwt.RegisteredClaims        // 标准JWT声明（包含过期时间、签发时间等）
}

// GenerateToken 生成包含用户信息的JWT令牌
// 参数:
//   - userID: 用户ID
//   - username: 用户名
//   - _: 预留参数，当前未使用
//
// 返回:
//   - 生成的JWT令牌字符串
//   - 可能的错误
func GenerateToken(userID uint, username string, _ string) (string, error) {
	// 获取JWT配置
	jwtConfig := config.GetJWTConfig()

	// 解析过期时间
	expDuration, err := time.ParseDuration(jwtConfig.ExpiresTime)
	if err != nil {
		return "", fmt.Errorf("解析过期时间失败: %w", err)
	}

	// 获取当前时间
	now := time.Now()

	// 创建自定义声明
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expDuration)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(now),                  // 签发时间
			NotBefore: jwt.NewNumericDate(now),                  // 生效时间
			Issuer:    jwtConfig.Issuer,                         // 签发者
			ID:        uuid.New().String(),                      // 唯一ID
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(jwtConfig.SecretKey))
	if err != nil {
		return "", fmt.Errorf("签名令牌失败: %w", err)
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌并提取其中的声明信息
// 参数:
//   - tokenString: JWT令牌字符串
//
// 返回:
//   - 解析出的CustomClaims指针
//   - 可能的错误，包括令牌过期、无效或未提供
func ParseToken(tokenString string) (*CustomClaims, error) {
	// 检查令牌是否为空
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	// 获取JWT配置
	jwtConfig := config.GetJWTConfig()

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		// 返回用于验证签名的密钥
		return []byte(jwtConfig.SecretKey), nil
	})

	// 处理解析错误
	if err != nil {
		// 特殊处理令牌过期错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// 提取并类型转换声明
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ValidateToken 验证JWT令牌有效性
func ValidateToken(tokenString string) (bool, error) {
	_, err := ParseToken(tokenString)
	if err != nil {
		return false, err
	}
	return true, nil
}

// RefreshToken 刷新JWT令牌
func RefreshToken(tokenString string) (string, error) {
	// 解析原令牌，即使已过期
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

	// 生成新令牌
	return GenerateToken(claims.UserID, claims.Username, "")
}

// parseTokenWithoutValidation 解析JWT令牌但不验证过期时间，用于令牌刷新
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
