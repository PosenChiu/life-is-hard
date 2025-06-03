package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"life-is-hard/internal/cache"
	"life-is-hard/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// helper to create a string of length n
func longString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func TestHashAndComparePassword(t *testing.T) {
	// success case
	hash, err := HashPassword("secret")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NoError(t, ComparePassword(hash, "secret"))

	// failure when password too long
	_, err = HashPassword(longString(80))
	require.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
	require.Error(t, ComparePassword(hash, "wrong"))
}

func TestAuthenticateUser(t *testing.T) {
	user := model.User{PasswordHash: func() string { h, _ := HashPassword("pwd"); return h }()}
	require.NoError(t, AuthenticateUser(context.Background(), user, "pwd"))
	err := AuthenticateUser(context.Background(), user, "wrong")
	require.EqualError(t, err, "invalid password")
}

func TestIssueAccessToken(t *testing.T) {
	user := model.User{ID: 1, IsAdmin: true}
	// secret missing
	os.Unsetenv("JWT_SECRET")
	_, err := IssueAccessToken(user, time.Minute)
	require.Error(t, err)

	// success path
	secret := "s3cr3t"
	os.Setenv("JWT_SECRET", secret)
	token, err := IssueAccessToken(user, time.Minute)
	require.NoError(t, err)
	claims, err := VerifyAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, 1, claims.UserID)
	require.True(t, claims.IsAdmin)
}

func TestIssueClientAccessToken(t *testing.T) {
	user := model.User{ID: 1}
	client := model.OAuthClient{ClientID: "cid", UserID: 1}

	os.Unsetenv("JWT_SECRET")
	_, err := IssueClientAccessToken(user, client, time.Minute)
	require.Error(t, err)

	secret := "sec"
	os.Setenv("JWT_SECRET", secret)
	client.UserID = 2
	_, err = IssueClientAccessToken(user, client, time.Minute)
	require.Error(t, err)

	client.UserID = 1
	token, err := IssueClientAccessToken(user, client, time.Minute)
	require.NoError(t, err)
	claims, err := VerifyAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, "cid", claims.ClientID)
}

func TestVerifyAccessToken(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	_, err := VerifyAccessToken("tok")
	require.Error(t, err)

	secret := "verify"
	os.Setenv("JWT_SECRET", secret)
	// token with different signing method
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 512)
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "1"}).SignedString(rsaKey)
	_, err = VerifyAccessToken(tok)
	require.Error(t, err)
	require.Error(t, err)

	// valid token
	good, _ := IssueAccessToken(model.User{ID: 9}, time.Minute)
	claims, err := VerifyAccessToken(good)
	require.NoError(t, err)
	require.Equal(t, 9, claims.UserID)
}

func TestIssueRefreshToken(t *testing.T) {
	ctx := context.Background()
	c := &cache.FakeCache{
		SetFn: func(_ context.Context, key string, value any, ttl time.Duration) *redis.StatusCmd {
			return redis.NewStatusResult("OK", nil)
		},
	}
	token, err := IssueRefreshToken(ctx, c, 1, "cid", true, time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	// cache.Set fail
	c.SetFn = func(_ context.Context, key string, value any, ttl time.Duration) *redis.StatusCmd {
		return redis.NewStatusResult("", errors.New("store fail"))
	}
	_, err = IssueRefreshToken(ctx, c, 1, "cid", true, time.Minute)
	require.Error(t, err)
}

func TestValidateRefreshToken(t *testing.T) {
	ctx := context.Background()
	data := RefreshTokenData{UserID: 1, ClientID: "cid", IsAdmin: true}
	bs, _ := json.Marshal(data)
	c := &cache.FakeCache{
		GetFn: func(_ context.Context, key string) *redis.StringCmd {
			return redis.NewStringResult(string(bs), nil)
		},
	}
	token := "abc"
	got, err := ValidateRefreshToken(ctx, c, token)
	require.NoError(t, err)
	require.Equal(t, data.UserID, got.UserID)

	// not found
	c.GetFn = func(_ context.Context, key string) *redis.StringCmd {
		return redis.NewStringResult("", redis.Nil)
	}
	_, err = ValidateRefreshToken(ctx, c, token)
	require.Error(t, err)

	// other error
	c.GetFn = func(_ context.Context, key string) *redis.StringCmd {
		return redis.NewStringResult("", errors.New("fail"))
	}
	_, err = ValidateRefreshToken(ctx, c, token)
	require.Error(t, err)

	// invalid json
	c.GetFn = func(_ context.Context, key string) *redis.StringCmd {
		return redis.NewStringResult("{", nil)
	}
	_, err = ValidateRefreshToken(ctx, c, token)
	require.Error(t, err)
}
