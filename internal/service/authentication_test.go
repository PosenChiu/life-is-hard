package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"life-is-hard/internal/cache"
	"life-is-hard/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func restoreGlobals() {
	bcryptGenerateFromPassword = bcrypt.GenerateFromPassword
	bcryptCompareHashAndPassword = bcrypt.CompareHashAndPassword
	randRead = rand.Read
	jsonMarshal = json.Marshal
	jsonUnmarshal = json.Unmarshal
	timeNow = time.Now
	parseWithClaims = jwt.ParseWithClaims
}

func TestHashPassword(t *testing.T) {
	t.Cleanup(restoreGlobals)
	pwd := "secret"
	hash, err := HashPassword(pwd)
	require.NoError(t, err)
	require.NotEqual(t, pwd, hash)
	require.NoError(t, ComparePassword(hash, pwd))

	bcryptGenerateFromPassword = func(_ []byte, _ int) ([]byte, error) {
		return nil, errors.New("gen")
	}
	_, err = HashPassword(pwd)
	require.Error(t, err)
}

func TestAuthenticateUser(t *testing.T) {
	t.Cleanup(restoreGlobals)
	hash, _ := HashPassword("pw")
	u := model.User{PasswordHash: hash}
	require.NoError(t, AuthenticateUser(context.Background(), u, "pw"))
	require.Error(t, AuthenticateUser(context.Background(), u, "bad"))
}

func TestIssueAccessToken(t *testing.T) {
	t.Cleanup(restoreGlobals)
	os.Unsetenv("JWT_SECRET")
	_, err := IssueAccessToken(model.User{}, time.Minute)
	require.Error(t, err)

	os.Setenv("JWT_SECRET", "s")
	tok, err := IssueAccessToken(model.User{ID: 5, IsAdmin: true}, time.Minute)
	require.NoError(t, err)
	claims := &CustomClaims{}
	_, err = jwt.ParseWithClaims(tok, claims, func(*jwt.Token) (any, error) { return []byte("s"), nil })
	require.NoError(t, err)
	require.Equal(t, 5, claims.UserID)
	require.True(t, claims.IsAdmin)
}

func TestIssueClientAccessToken(t *testing.T) {
	t.Cleanup(restoreGlobals)
	user := model.User{ID: 1}
	client := model.OAuthClient{ClientID: "c", UserID: 1}

	os.Unsetenv("JWT_SECRET")
	_, err := IssueClientAccessToken(user, client, time.Minute)
	require.Error(t, err)

	os.Setenv("JWT_SECRET", "s")
	_, err = IssueClientAccessToken(model.User{ID: 2}, client, time.Minute)
	require.Error(t, err)

	tok, err := IssueClientAccessToken(user, client, time.Hour)
	require.NoError(t, err)
	c := &CustomClaims{}
	_, err = jwt.ParseWithClaims(tok, c, func(*jwt.Token) (any, error) { return []byte("s"), nil })
	require.NoError(t, err)
	require.Equal(t, "c", c.ClientID)
}

func TestVerifyAccessToken(t *testing.T) {
	t.Cleanup(restoreGlobals)
	os.Unsetenv("JWT_SECRET")
	_, err := VerifyAccessToken("abc")
	require.Error(t, err)

	os.Setenv("JWT_SECRET", "s")
	_, err = VerifyAccessToken("invalid")
	require.Error(t, err)

	tokNone, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"foo": "bar"}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	_, err = VerifyAccessToken(tokNone)
	require.Error(t, err)

	parseWithClaims = func(s string, c jwt.Claims, k jwt.Keyfunc, opts ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{Claims: jwt.MapClaims{}, Valid: false}, nil
	}
	_, err = VerifyAccessToken("whatever")
	require.Error(t, err)

	parseWithClaims = jwt.ParseWithClaims
	tok, _ := IssueAccessToken(model.User{ID: 3}, time.Minute)
	claims, err := VerifyAccessToken(tok)
	require.NoError(t, err)
	require.Equal(t, 3, claims.UserID)
}

func TestIssueRefreshToken(t *testing.T) {
	t.Cleanup(restoreGlobals)
	ctx := context.Background()
	c := &cache.FakeCache{}

	randRead = func([]byte) (int, error) { return 0, errors.New("rand") }
	_, err := IssueRefreshToken(ctx, c, 1, "cli", false, time.Second)
	require.Error(t, err)

	randRead = rand.Read
	jsonMarshal = func(any) ([]byte, error) { return nil, errors.New("json") }
	_, err = IssueRefreshToken(ctx, c, 1, "cli", false, time.Second)
	require.Error(t, err)

	jsonMarshal = json.Marshal
	c.SetFn = func(context.Context, string, any, time.Duration) *redis.StatusCmd {
		return redis.NewStatusResult("", errors.New("set"))
	}
	_, err = IssueRefreshToken(ctx, c, 1, "cli", false, time.Second)
	require.Error(t, err)

	var storedKey string
	var storedVal []byte
	c.SetFn = func(_ context.Context, key string, val any, _ time.Duration) *redis.StatusCmd {
		storedKey = key
		storedVal = val.([]byte)
		return redis.NewStatusResult("OK", nil)
	}
	tok, err := IssueRefreshToken(ctx, c, 1, "cli", true, time.Second)
	require.NoError(t, err)
	require.Contains(t, storedKey, tok)
	decoded, _ := base64.RawURLEncoding.DecodeString(tok)
	require.Len(t, decoded, 32)
	var d RefreshTokenData
	require.NoError(t, json.Unmarshal(storedVal, &d))
	require.Equal(t, 1, d.UserID)
	require.Equal(t, "cli", d.ClientID)
	require.True(t, d.IsAdmin)
}

func TestValidateRefreshToken(t *testing.T) {
	t.Cleanup(restoreGlobals)
	ctx := context.Background()
	c := &cache.FakeCache{}

	c.GetFn = func(context.Context, string) *redis.StringCmd {
		return redis.NewStringResult("", redis.Nil)
	}
	_, err := ValidateRefreshToken(ctx, c, "tok")
	require.Error(t, err)

	c.GetFn = func(context.Context, string) *redis.StringCmd {
		return redis.NewStringResult("", errors.New("get"))
	}
	_, err = ValidateRefreshToken(ctx, c, "tok")
	require.Error(t, err)

	c.GetFn = func(context.Context, string) *redis.StringCmd {
		return redis.NewStringResult("bad", nil)
	}
	jsonUnmarshal = func([]byte, any) error { return errors.New("unmarshal") }
	_, err = ValidateRefreshToken(ctx, c, "tok")
	require.Error(t, err)

	jsonUnmarshal = json.Unmarshal
	dataBytes, _ := json.Marshal(RefreshTokenData{UserID: 2, ClientID: "c", IsAdmin: true})
	c.GetFn = func(context.Context, string) *redis.StringCmd {
		return redis.NewStringResult(string(dataBytes), nil)
	}
	data, err := ValidateRefreshToken(ctx, c, "tok")
	require.NoError(t, err)
	require.Equal(t, 2, data.UserID)
	require.Equal(t, "c", data.ClientID)
	require.True(t, data.IsAdmin)
}
