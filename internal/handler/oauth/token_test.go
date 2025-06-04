package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"
	"life-is-hard/internal/model"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// fakeUserRow implements pgx.Row for user queries
type fakeUserRow struct {
	user *model.User
	err  error
}

func (r *fakeUserRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	u := r.user
	*dest[0].(*int) = u.ID
	*dest[1].(*string) = u.Name
	*dest[2].(*string) = u.Email
	*dest[3].(*string) = u.PasswordHash
	*dest[4].(*time.Time) = u.CreatedAt
	*dest[5].(*bool) = u.IsAdmin
	return nil
}

// fakeClientRow implements pgx.Row for oauth client queries
type fakeClientRow struct {
	client *model.OAuthClient
	err    error
}

func (r *fakeClientRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	c := r.client
	*dest[0].(*string) = c.ClientID
	*dest[1].(*string) = c.ClientSecret
	*dest[2].(*int) = c.UserID
	*dest[3].(*[]string) = c.GrantTypes
	*dest[4].(*time.Time) = c.CreatedAt
	*dest[5].(*time.Time) = c.UpdatedAt
	return nil
}

// helper to create echo context with form body and Authorization header
func newCtx(e *echo.Echo, form string, auth string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestTokenHandler(t *testing.T) {
	e := echo.New()
	now := time.Now()
	hashed, _ := service.HashPassword("pw")
	user := &model.User{ID: 1, Name: "u", Email: "e", PasswordHash: hashed, CreatedAt: now}
	client := &model.OAuthClient{ClientID: "cid", ClientSecret: "sec", UserID: 1, GrantTypes: []string{"password", "client_credentials", "refresh_token"}, CreatedAt: now, UpdatedAt: now}

	validAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("cid:sec"))

	t.Run("bind error", func(t *testing.T) {
		ctx, rec := newCtx(e, "bad%", validAuth)
		err := TokenHandler(&database.FakeDB{}, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid request payload")
	})

	t.Run("invalid auth prefix", func(t *testing.T) {
		ctx, rec := newCtx(e, "grant_type=password", "")
		err := TokenHandler(&database.FakeDB{}, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid authorization header")
	})

	t.Run("decode error", func(t *testing.T) {
		ctx, rec := newCtx(e, "grant_type=password", "Basic !!!")
		err := TokenHandler(&database.FakeDB{}, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid authorization header")
	})

	t.Run("split error", func(t *testing.T) {
		bad := base64.StdEncoding.EncodeToString([]byte("cid-sec"))
		ctx, rec := newCtx(e, "grant_type=password", "Basic "+bad)
		err := TokenHandler(&database.FakeDB{}, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid client", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			return &fakeClientRow{err: errors.New("no")}
		}}
		ctx, rec := newCtx(e, "grant_type=password", validAuth)
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("unauthorized grant", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(context.Context, string, ...any) pgx.Row {
			return &fakeClientRow{client: &model.OAuthClient{ClientID: "cid", ClientSecret: "sec", GrantTypes: []string{"client_credentials"}, CreatedAt: now, UpdatedAt: now}}
		}}
		ctx, rec := newCtx(e, "grant_type=password", validAuth)
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("password user not found", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			if strings.Contains(q, "FROM oauth_clients") {
				return &fakeClientRow{client: client}
			}
			return &fakeUserRow{err: errors.New("no user")}
		}}
		ctx, rec := newCtx(e, "grant_type=password&username=x&password=pw", validAuth)
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("password auth fail", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			if strings.Contains(q, "FROM oauth_clients") {
				return &fakeClientRow{client: client}
			}
			return &fakeUserRow{user: user}
		}}
		ctx, rec := newCtx(e, "grant_type=password&username=u&password=bad", validAuth)
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("password issue access token fail", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			if strings.Contains(q, "FROM oauth_clients") {
				return &fakeClientRow{client: client}
			}
			return &fakeUserRow{user: user}
		}}
		ctx, rec := newCtx(e, "grant_type=password&username=u&password=pw", validAuth)
		t.Setenv("JWT_SECRET", "")
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "failed to issue token")
	})

	t.Run("password issue refresh token fail", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			if strings.Contains(q, "FROM oauth_clients") {
				return &fakeClientRow{client: client}
			}
			return &fakeUserRow{user: user}
		}}
		cch := &cache.FakeCache{SetFn: func(context.Context, string, any, time.Duration) *redis.StatusCmd {
			return redis.NewStatusResult("", errors.New("set"))
		}}
		ctx, rec := newCtx(e, "grant_type=password&username=u&password=pw", validAuth)
		t.Setenv("JWT_SECRET", "s")
		err := TokenHandler(db, cch)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "failed to issue refresh token")
	})

	t.Run("password success", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			if strings.Contains(q, "FROM oauth_clients") {
				return &fakeClientRow{client: client}
			}
			return &fakeUserRow{user: user}
		}}
		cch := &cache.FakeCache{SetFn: func(context.Context, string, any, time.Duration) *redis.StatusCmd {
			return redis.NewStatusResult("OK", nil)
		}}
		ctx, rec := newCtx(e, "grant_type=password&username=u&password=pw", validAuth)
		t.Setenv("JWT_SECRET", "s")
		err := TokenHandler(db, cch)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "access_token")
		require.Contains(t, rec.Body.String(), "refresh_token")
	})

	t.Run("client creds owner error", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			if strings.Contains(q, "FROM oauth_clients") {
				return &fakeClientRow{client: client}
			}
			return &fakeUserRow{err: errors.New("owner")}
		}}
		ctx, rec := newCtx(e, "grant_type=client_credentials", validAuth)
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "failed to retrieve client owner")
	})

	t.Run("client creds issue token fail", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			if strings.Contains(q, "FROM oauth_clients") {
				return &fakeClientRow{client: client}
			}
			return &fakeUserRow{user: user}
		}}
		ctx, rec := newCtx(e, "grant_type=client_credentials", validAuth)
		t.Setenv("JWT_SECRET", "")
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("client creds success", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			if strings.Contains(q, "FROM oauth_clients") {
				return &fakeClientRow{client: client}
			}
			return &fakeUserRow{user: user}
		}}
		ctx, rec := newCtx(e, "grant_type=client_credentials", validAuth)
		t.Setenv("JWT_SECRET", "s")
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "access_token")
	})

	t.Run("refresh token invalid", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			return &fakeClientRow{client: client}
		}}
		cch := &cache.FakeCache{GetFn: func(context.Context, string) *redis.StringCmd {
			return redis.NewStringResult("", redis.Nil)
		}}
		ctx, rec := newCtx(e, "grant_type=refresh_token&refresh_token=tok", validAuth)
		err := TokenHandler(db, cch)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("refresh token issue access token fail", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			return &fakeClientRow{client: client}
		}}
		dataBytes, _ := json.Marshal(service.RefreshTokenData{UserID: 1, ClientID: "cid"})
		cch := &cache.FakeCache{GetFn: func(context.Context, string) *redis.StringCmd {
			return redis.NewStringResult(string(dataBytes), nil)
		}}
		ctx, rec := newCtx(e, "grant_type=refresh_token&refresh_token=tok", validAuth)
		t.Setenv("JWT_SECRET", "")
		err := TokenHandler(db, cch)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("refresh token success", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			return &fakeClientRow{client: client}
		}}
		dataBytes, _ := json.Marshal(service.RefreshTokenData{UserID: 1, ClientID: "cid"})
		cch := &cache.FakeCache{GetFn: func(context.Context, string) *redis.StringCmd {
			return redis.NewStringResult(string(dataBytes), nil)
		}}
		ctx, rec := newCtx(e, "grant_type=refresh_token&refresh_token=tok", validAuth)
		t.Setenv("JWT_SECRET", "s")
		err := TokenHandler(db, cch)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "access_token")
		require.Contains(t, rec.Body.String(), "refresh_token")
	})

	t.Run("unsupported grant type", func(t *testing.T) {
		db := &database.FakeDB{QueryRowFn: func(ctx context.Context, q string, args ...any) pgx.Row {
			return &fakeClientRow{client: &model.OAuthClient{ClientID: "cid", ClientSecret: "sec", GrantTypes: []string{"foo"}, CreatedAt: now, UpdatedAt: now}}
		}}
		ctx, rec := newCtx(e, "grant_type=foo", validAuth)
		err := TokenHandler(db, &cache.FakeCache{})(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
