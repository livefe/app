package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// 定义错误常量
var (
	// ErrInvalidBase64Data 表示提供的Base64数据无效
	ErrInvalidBase64Data = errors.New("提供的Base64图片数据无效")
	// ErrUnsupportedImageFormat 表示不支持的图片格式
	ErrUnsupportedImageFormat = errors.New("不支持的图片格式")
)

// ParseBase64Image 解析Base64编码的图片数据
// 返回一个io.Reader用于读取图片数据、文件名、图片大小和可能的错误
func ParseBase64Image(base64Data string) (io.Reader, string, int64, error) {
	// 检查输入是否为空
	if base64Data == "" {
		return nil, "", 0, ErrInvalidBase64Data
	}

	// 处理可能的Data URL格式 (如 "data:image/jpeg;base64,/9j/4AAQSkZ...")
	var encodedData string
	var imageType string

	if strings.Contains(base64Data, "base64,") {
		// 分离MIME类型和Base64数据
		parts := strings.SplitN(base64Data, "base64,", 2)
		if len(parts) != 2 {
			return nil, "", 0, ErrInvalidBase64Data
		}

		// 提取MIME类型
		mimeType := parts[0]
		encodedData = parts[1]

		// 从MIME类型中提取图片格式
		if strings.Contains(mimeType, "image/jpeg") || strings.Contains(mimeType, "image/jpg") {
			imageType = "jpg"
		} else if strings.Contains(mimeType, "image/png") {
			imageType = "png"
		} else if strings.Contains(mimeType, "image/gif") {
			imageType = "gif"
		} else if strings.Contains(mimeType, "image/webp") {
			imageType = "webp"
		} else {
			// 默认为JPEG
			imageType = "jpg"
		}
	} else {
		// 假设直接是Base64编码的数据
		encodedData = base64Data
		// 默认为JPEG格式
		imageType = "jpg"

		// 尝试通过检查Base64数据的前几个字节来确定图片类型
		// 这种方法不是100%可靠，但对于大多数常见格式有效
		if len(encodedData) > 10 {
			prefix := encodedData[:10]
			if strings.Contains(prefix, "/9j/") {
				// JPEG的特征
				imageType = "jpg"
			} else if strings.Contains(prefix, "iVBORw0") {
				// PNG的特征
				imageType = "png"
			} else if strings.Contains(prefix, "R0lGOD") {
				// GIF的特征
				imageType = "gif"
			} else if strings.Contains(prefix, "UklGR") {
				// WEBP的特征
				imageType = "webp"
			}
		}
	}

	// 解码Base64数据
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, "", 0, fmt.Errorf("解码Base64数据失败: %w", err)
	}

	// 检查解码后的数据大小
	size := int64(len(decoded))
	if size == 0 {
		return nil, "", 0, ErrInvalidBase64Data
	}

	// 创建一个唯一的文件名
	timestamp := time.Now().UnixNano() / 1000000 // 毫秒级时间戳
	filename := fmt.Sprintf("image_%d.%s", timestamp, imageType)

	// 创建一个bytes.Reader作为io.Reader返回
	reader := bytes.NewReader(decoded)

	return reader, filename, size, nil
}
