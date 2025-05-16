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

type CustomClaims struct {
	UserID   int    `json:"user_id,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
	jwt.RegisteredClaims
}

type RefreshTokenData struct {
	UserID   int    `json:"user_id"`
	ClientID string `json:"client_id"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
}

func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}

func ComparePassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func AuthenticateUser(ctx context.Context, user model.User, password string) error {
	if err := ComparePassword(user.PasswordHash, password); err != nil {
		return errors.New("invalid password")
	}
	return nil
}

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

func IssueClientAccessToken(user model.User, client model.OAuthClient, ttl time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}
	if user.ID != client.UserID {
		return "", fmt.Errorf("user %d is not the owner of client %s", user.ID, client.ClientID)
	}
	now := time.Now()
	claims := CustomClaims{
		UserID:   user.ID,
		ClientID: client.ClientID,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(client.ClientID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

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

func IssueRefreshToken(ctx context.Context, rdb *redis.Client, userID int, clientID string, isAdmin bool, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	data := RefreshTokenData{UserID: userID, ClientID: clientID, IsAdmin: isAdmin}
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
