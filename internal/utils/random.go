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

	// 计算需要多少随机字节来确保均匀分布
	// 我们需要足够的随机性来覆盖字符集大小
	// 使用比特位计算来确保均匀分布
	maxByte := 255
	neededBytes := length * ((maxByte / charsetLen) + 1)
	buf := make([]byte, neededBytes)

	_, err := rand.Read(buf)
	if err != nil {
		// 如果安全随机数生成失败，回退到不太安全的方法
		// 使用math/rand作为备用，但增加熵源
		source := mathrand.NewSource(time.Now().UnixNano())
		r := mathrand.New(source)
		for i := 0; i < length; i++ {
			result[i] = charset[r.Intn(charsetLen)]
		}
		return string(result)
	}

	// 将随机字节转换为指定字符集中的字符
	// 使用取模偏差修正算法确保均匀分布
	bufIndex := 0
	for i := 0; i < length; i++ {
		// 寻找一个在有效范围内的随机值
		// 这样可以避免模运算导致的分布不均
		for bufIndex < len(buf) {
			// 计算阈值，确保均匀分布
			threshold := 256 - (256 % charsetLen)
			if int(buf[bufIndex]) < threshold {
				result[i] = charset[int(buf[bufIndex])%charsetLen]
				bufIndex++
				break
			}
			bufIndex++
		}

		// 如果用完了随机字节但还没完成，回退到简单方法
		if bufIndex >= len(buf) && i < length-1 {
			source := mathrand.NewSource(time.Now().UnixNano())
			r := mathrand.New(source)
			for j := i + 1; j < length; j++ {
				result[j] = charset[r.Intn(charsetLen)]
			}
			break
		}
	}

	return string(result)
}
