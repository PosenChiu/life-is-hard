// File: internal/service/password.go
package service

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 接收明文密碼，回傳 bcrypt 哈希字串
func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}

// ComparePassword 比對明文密碼與 bcrypt 哈希，成功回傳 nil，失敗則回傳錯誤
func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
