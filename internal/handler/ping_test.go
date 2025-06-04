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

func TestPingHandler(t *testing.T) {
	e := echo.New()

	t.Run("db unhealthy", func(t *testing.T) {
		db := &database.FakeDB{
			PingFn:     func(ctx context.Context) error { return errors.New("fail") },
			QueryRowFn: func(context.Context, string, ...any) pgx.Row { return nil },
		}
		cch := &cache.FakeCache{}
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		err := PingHandler(db, cch)(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "database unhealthy")
	})

	t.Run("cache unhealthy", func(t *testing.T) {
		dbCalled := false
		db := &database.FakeDB{
			PingFn:     func(ctx context.Context) error { dbCalled = true; return nil },
			QueryRowFn: func(context.Context, string, ...any) pgx.Row { return nil },
		}
		cch := &cache.FakeCache{SetFn: func(ctx context.Context, key string, val any, exp time.Duration) *redis.StatusCmd {
			return redis.NewStatusResult("", errors.New("set"))
		}}
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		err := PingHandler(db, cch)(ctx)
		require.NoError(t, err)
		require.True(t, dbCalled)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "cache unhealthy")
	})

	t.Run("ok", func(t *testing.T) {
		dbCalled := false
		cacheCalled := false
		db := &database.FakeDB{
			PingFn:     func(ctx context.Context) error { dbCalled = true; return nil },
			QueryRowFn: func(context.Context, string, ...any) pgx.Row { return nil },
		}
		cch := &cache.FakeCache{SetFn: func(ctx context.Context, key string, val any, exp time.Duration) *redis.StatusCmd {
			cacheCalled = true
			return redis.NewStatusResult("OK", nil)
		}}
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		err := PingHandler(db, cch)(ctx)
		require.NoError(t, err)
		require.True(t, dbCalled)
		require.True(t, cacheCalled)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "pong")
	})
}
