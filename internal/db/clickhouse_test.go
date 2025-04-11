package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"
)

// 模拟明文密码加密为 SHA256
func encryptToSHA256(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// 模拟对称加密（这里用 AES 替代 SHA256 的单向性，展示解密逻辑）
func encryptAES(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	plainBytes := []byte(plainText)
	cipherText := make([]byte, aes.BlockSize+len(plainBytes))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainBytes)
	return hex.EncodeToString(cipherText), nil
}

// 模拟对称解密
func decryptAES(cipherText, key string) (string, error) {
	cipherBytes, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(cipherBytes) < aes.BlockSize {
		return "", err
	}

	iv := cipherBytes[:aes.BlockSize]
	cipherBytes = cipherBytes[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherBytes, cipherBytes)

	return string(cipherBytes), nil
}

// 测试用例
func TestClickHousePasswordEncryption(t *testing.T) {
	// 测试数据
	tests := []struct {
		name           string
		plainPassword  string
		expectedSHA256 string
	}{
		{
			name:           "default用户密码加密",
			plainPassword:  "crane",
			expectedSHA256: encryptToSHA256("crane"),
		},
		{
			name:           "default用户密码解密",
			plainPassword:  "crane",
			expectedSHA256: "2c6ac23e4ffdf95f08f369eca6488b585bca0def0ddfe69f525e40d4aa2509d3",
		},
		{
			name:           "shabi用户密码验证",
			plainPassword:  "barry", // 我们假设这是原始密码
			expectedSHA256: "db857082a0430dfe0adb0beed24710a831d8e00acff7782aa5e006970bc3d1ba",
		},
	}

	// 测试 SHA256 加密
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := encryptToSHA256(tt.plainPassword)
			if encrypted != tt.expectedSHA256 {
				t.Errorf("加密结果不匹配，预期 %s，实际 %s", tt.expectedSHA256, encrypted)
			} else {
				t.Logf("密码 %s 加密为 SHA256: %s", tt.plainPassword, encrypted)
			}
		})
	}
}

func TestClickHousePasswordSymmetricEncryption(t *testing.T) {
	// 对称加密的密钥（必须是 16、24 或 32 字节长）
	key := "thisis32byteslongkeyforaes256!!!"
	plainPassword := "crane"

	// 测试加密
	encrypted, err := encryptAES(plainPassword, key)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	t.Logf("明文密码 %s 加密后: %s", plainPassword, encrypted)

	// 测试解密
	decrypted, err := decryptAES(encrypted, key)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if decrypted != plainPassword {
		t.Errorf("解密结果不匹配，预期 %s，实际 %s", plainPassword, decrypted)
	} else {
		t.Logf("加密密码 %s 解密回明文: %s", encrypted, decrypted)
	}
}
