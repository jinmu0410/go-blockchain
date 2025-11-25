package bip

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

// KeyStore 密钥存储接口
type KeyStore interface {
	// SaveKey 保存密钥（加密）
	SaveKey(ctx context.Context, address string, encryptedKey []byte) error

	// GetKey 获取密钥（解密）
	GetKey(ctx context.Context, address string) ([]byte, error)

	// DeleteKey 删除密钥
	DeleteKey(ctx context.Context, address string) error
}

// EncryptPrivateKey 加密私钥
func EncryptPrivateKey(privateKey []byte, password string) ([]byte, error) {
	// 使用 scrypt 派生密钥
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	// 使用 AES-256-GCM 加密
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, privateKey, nil)

	// 组合 salt + ciphertext
	encrypted := append(salt, ciphertext...)
	return encrypted, nil
}

// DecryptPrivateKey 解密私钥
func DecryptPrivateKey(encrypted []byte, password string) ([]byte, error) {
	if len(encrypted) < 32 {
		return nil, fmt.Errorf("invalid encrypted data length")
	}

	// 提取 salt 和 ciphertext
	salt := encrypted[:32]
	ciphertext := encrypted[32:]

	// 使用 scrypt 派生密钥
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	// 使用 AES-256-GCM 解密
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("invalid ciphertext length")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// DeriveKeyFromPassword 从密码派生密钥（PBKDF2）
func DeriveKeyFromPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
}

// HashAddress 哈希地址（用于存储索引）
func HashAddress(address string) string {
	hash := sha256.Sum256([]byte(address))
	return hex.EncodeToString(hash[:])
}

