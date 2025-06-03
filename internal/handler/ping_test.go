package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newPingCtx() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestPingHandler(t *testing.T) {
	// db error
	ctx, rec := newPingCtx()
	db := &database.FakeDB{
		QueryRowFn: func(context.Context, string, ...any) pgx.Row { return nil },
		PingFn:     func(context.Context) error { return errors.New("db") },
	}
	h := PingHandler(db, &cache.FakeCache{})
	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// cache error
	ctx, rec = newPingCtx()
	c := &cache.FakeCache{SetFn: func(context.Context, string, any, time.Duration) *redis.StatusCmd {
		return redis.NewStatusResult("", errors.New("cache"))
	}}
	db = &database.FakeDB{
		QueryRowFn: func(context.Context, string, ...any) pgx.Row { return nil },
		PingFn:     func(context.Context) error { return nil },
	}
	h = PingHandler(db, c)
	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// success
	ctx, rec = newPingCtx()
	c = &cache.FakeCache{SetFn: func(context.Context, string, any, time.Duration) *redis.StatusCmd {
		return redis.NewStatusResult("OK", nil)
	}}
	h = PingHandler(db, c)
	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "pong")
}
