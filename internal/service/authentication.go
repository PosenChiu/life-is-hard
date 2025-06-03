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

	"life-is-hard/internal/cache"
	"life-is-hard/internal/model"

	"github.com/redis/go-redis/v9"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	bcryptGenerateFromPassword   = bcrypt.GenerateFromPassword
	bcryptCompareHashAndPassword = bcrypt.CompareHashAndPassword
	randRead                     = rand.Read
	jsonMarshal                  = json.Marshal
	jsonUnmarshal                = json.Unmarshal
	timeNow                      = time.Now
	parseWithClaims              = jwt.ParseWithClaims
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
	hashBytes, err := bcryptGenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}

func ComparePassword(hash string, password string) error {
	return bcryptCompareHashAndPassword([]byte(hash), []byte(password))
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
	now := timeNow()
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
	now := timeNow()
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
	token, err := parseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
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

func IssueRefreshToken(ctx context.Context, cache cache.Cache, userID int, clientID string, isAdmin bool, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := randRead(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	data := RefreshTokenData{UserID: userID, ClientID: clientID, IsAdmin: isAdmin}
	bytesData, err := jsonMarshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh token data: %w", err)
	}
	key := fmt.Sprintf("refresh_token:%s", token)
	if err := cache.Set(ctx, key, bytesData, ttl).Err(); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}
	return token, nil
}

func ValidateRefreshToken(ctx context.Context, cache cache.Cache, token string) (*RefreshTokenData, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	val, err := cache.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("refresh token not found or expired")
		}
		return nil, fmt.Errorf("failed to retrieve refresh token: %w", err)
	}
	var data RefreshTokenData
	if err := jsonUnmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to parse refresh token data: %w", err)
	}
	return &data, nil
}
