// File: internal/service/authentication.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"life-is-hard/internal/model"

	"github.com/redis/go-redis/v9"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// CustomClaims 定義 JWT 負載內容
// 包含 userID, clientID, isAdmin 以及標準註冊聲明
// 可以用於 access token
type CustomClaims struct {
	UserID   int  `json:"user_id,omitempty"`
	ClientID int  `json:"client_id,omitempty"`
	IsAdmin  bool `json:"is_admin,omitempty"`
	jwt.RegisteredClaims
}

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

// AuthenticateUser 根據使用者結構和明文密碼驗證，成功回傳使用者
func AuthenticateUser(ctx context.Context, user model.User, password string) (*model.User, error) {
	if user.PasswordHash == nil {
		if password == "" {
			return &user, nil
		}
		return nil, errors.New("invalid password")
	}
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
		UserID:  user.ID,
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

// IssueClientAccessToken 依據使用者與 client 資訊產生 JWT
func IssueClientAccessToken(user model.User, client model.OAuthClient, ttl time.Duration) (string, error) {
	if user.ID != client.OwnerID {
		return "", fmt.Errorf("user %d is not the owner of client %d", user.ID, client.ID)
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}

	now := time.Now()
	claims := CustomClaims{
		UserID:   user.ID,
		ClientID: client.ID,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(client.ID),
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

// RefreshTokenData 定義儲存在 Redis 中的資料結構
// 包含 userID, clientID, 以及 scope
type RefreshTokenData struct {
	UserID   int    `json:"user_id"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
}

// IssueRefreshToken 產生並儲存 refresh token
func IssueRefreshToken(ctx context.Context, rdb *redis.Client, userID int, clientID string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	data := RefreshTokenData{UserID: userID, ClientID: clientID}
	bytesData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh token data: %w", err)
	}
	key := fmt.Sprintf("refresh_token:%s", token)
	if err := rdb.Set(ctx, key, bytesData, ttl).Err(); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}
	return token, nil
}

// ValidateRefreshToken 驗證並讀取儲存在 Redis 的 refresh token
func ValidateRefreshToken(ctx context.Context, rdb *redis.Client, token string) (*RefreshTokenData, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("refresh token not found or expired")
		}
		return nil, fmt.Errorf("failed to retrieve refresh token: %w", err)
	}
	var data RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to parse refresh token data: %w", err)
	}
	return &data, nil
}
