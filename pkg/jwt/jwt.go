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

// 定义错误类型
var (
	ErrTokenExpired     = errors.New(constant.ErrTokenExpired)
	ErrTokenInvalid     = errors.New(constant.ErrTokenInvalid)
	ErrTokenNotProvided = errors.New(constant.ErrTokenNotProvided)
)

// CustomClaims 自定义JWT声明结构体
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, username string, _ string) (string, error) {
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
			ExpiresAt: jwt.NewNumericDate(now.Add(expDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),         // 添加NotBefore声明
			Issuer:    jwtConfig.Issuer,
			ID:        uuid.New().String(),             // 添加唯一ID
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

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*CustomClaims, error) {
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	jwtConfig := config.GetJWTConfig()

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
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

	// 验证令牌有效性
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// 提取声明
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

// parseTokenWithoutValidation 解析JWT令牌但不验证过期时间
func parseTokenWithoutValidation(tokenString string) (*CustomClaims, error) {
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	jwtConfig := config.GetJWTConfig()

	// 解析令牌但不验证过期时间
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
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

	// 提取声明
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
