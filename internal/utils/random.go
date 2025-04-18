package utils

import (
	"crypto/rand"
	mathrand "math/rand"
	"time"

	"app/pkg/logger"
)

// GenerateRandomDigits 生成指定长度的随机数字字符串
// 使用crypto/rand生成安全随机数，如果失败则回退到math/rand
func GenerateRandomDigits(length int) string {
	const digits = "0123456789"
	return GenerateRandomString(length, digits)
}

// GenerateRandomString 生成指定长度的随机字符串
// 从给定的字符集中随机选择字符
// 使用crypto/rand生成安全随机数，如果失败则回退到math/rand
func GenerateRandomString(length int, charset string) string {
	result := make([]byte, length)

	// 读取随机字节
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		// 如果安全随机数生成失败，回退到不太安全的方法
		logger.WithError(err).Warn("安全随机数生成失败，使用备用方法")
		// 使用math/rand作为备用
		r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
		for i := 0; i < length; i++ {
			result[i] = charset[r.Intn(len(charset))]
		}
		return string(result)
	}

	// 将随机字节转换为指定字符集中的字符
	for i := 0; i < length; i++ {
		result[i] = charset[int(buf[i])%len(charset)]
	}

	return string(result)
}
