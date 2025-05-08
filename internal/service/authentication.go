// File: internal/service/authentication.go
package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"life-is-hard/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 定義 JWT 負載內容
type CustomClaims struct {
	ID      int  `json:"id"`
	IsAdmin bool `json:"is_admin"`
	jwt.RegisteredClaims
}

// AuthenticateUser 根據使用者結構和明文密碼驗證，成功回傳使用者
func AuthenticateUser(ctx context.Context, user model.User, password string) (*model.User, error) {
	// 如果資料表未存密碼 (PasswordHash 為 nil)
	if user.PasswordHash == nil {
		if password == "" {
			return &user, nil
		}
		return nil, errors.New("invalid password")
	}
	// 使用 ComparePassword 進行 bcrypt 比對
	if err := ComparePassword(*user.PasswordHash, password); err != nil {
		return nil, errors.New("invalid password")
	}
	return &user, nil
}

// IssueAccessToken 依據使用者資訊與 TTL 產生 JWT
func IssueAccessToken(user model.User, ttl time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}

	now := time.Now()
	claims := CustomClaims{
		ID:      user.ID,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// VerifyAccessToken 驗證並解析 JWT 令牌
func VerifyAccessToken(tokenString string) (*CustomClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
