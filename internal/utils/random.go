package utils

import (
	"crypto/rand"
	mathrand "math/rand"
	"time"
)

// GenerateRandomDigits 生成指定长度的随机数字字符串
func GenerateRandomDigits(length int) string {
	const digits = "0123456789"
	return GenerateRandomString(length, digits)
}

// GenerateRandomString 生成指定长度的随机字符串，使用安全随机数生成器，失败时回退到math/rand
// 优化后的版本提高了随机性分布和安全性
func GenerateRandomString(length int, charset string) string {
	result := make([]byte, length)
	charsetLen := len(charset)

	// 使用crypto/rand生成随机字节
	_, err := rand.Read(result)
	if err != nil {
		// 如果安全随机数生成失败，回退到不太安全的方法
		source := mathrand.NewSource(time.Now().UnixNano())
		r := mathrand.New(source)
		for i := 0; i < length; i++ {
			result[i] = charset[r.Intn(charsetLen)]
		}
		return string(result)
	}

	// 将随机字节映射到字符集
	for i := 0; i < length; i++ {
		// 使用随机字节对字符集长度取模，获取字符集中的索引
		// 这种方法在字符集长度不是2的幂次时会有轻微的分布不均，但对大多数应用场景足够
		result[i] = charset[uint8(result[i])%uint8(charsetLen)]
	}

	return string(result)

}
