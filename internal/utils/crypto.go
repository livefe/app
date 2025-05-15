package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var (
	// ErrInvalidKey 表示提供的密钥无效
	ErrInvalidKey = errors.New("提供的密钥无效")
	// ErrInvalidData 表示提供的数据无效
	ErrInvalidData = errors.New("提供的数据无效")
	// ErrDecryptionFailed 表示解密操作失败
	ErrDecryptionFailed = errors.New("解密操作失败")
)

// EncryptAES 使用AES-GCM模式加密数据
// key 必须是32字节长度(AES-256)
// 返回base64编码的加密数据
func EncryptAES(plaintext []byte, key []byte) (string, error) {
	if len(key) == 0 {
		return "", ErrInvalidKey
	}

	// 使用SHA-256确保密钥长度为32字节
	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return "", err
	}

	// GCM模式提供认证加密
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 创建随机数
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 返回base64编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES 解密使用EncryptAES加密的数据
// key必须与加密时使用的相同
// 返回解密后的原始数据
func DecryptAES(encryptedData string, key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrInvalidKey
	}

	// 解码base64数据
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, ErrInvalidData
	}

	// 使用SHA-256确保密钥长度为32字节
	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}

	// GCM模式提供认证解密
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 检查数据长度是否有效
	if len(ciphertext) < gcm.NonceSize() {
		return nil, ErrInvalidData
	}

	// 提取nonce
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// HashPassword 对密码进行哈希处理
// 返回密码的SHA-256哈希值的base64编码
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}
