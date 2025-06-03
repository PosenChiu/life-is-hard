package main

import (
	"context"
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"life-is-hard/internal/cache"
	"life-is-hard/internal/database"
)

func restoreGlobals() {
	newPgxPool = database.NewPgxPool
	newRedisClient = cache.NewRedisClient
	runMigrationsFn = database.RunMigrations
	startServer = func(e *echo.Echo, addr string) error { return e.Start(addr) }
	exitFunc = func(code int) {}
}

func TestCustomValidator(t *testing.T) {
	cv := &CustomValidator{validator: validator.New()}
	type s struct {
		Name string `validate:"required"`
	}
	require.NoError(t, cv.Validate(&s{Name: "ok"}))
	require.Error(t, cv.Validate(&s{}))
}

func TestRunSuccess(t *testing.T) {
	t.Cleanup(restoreGlobals)
	called := make(map[string]bool)
	newPgxPool = func(ctx context.Context, url string) (database.DB, error) {
		called["pgx"] = true
		return &database.FakeDB{CloseFn: func() { called["dbClose"] = true }}, nil
	}
	newRedisClient = func(addr, pwd string, db int) (cache.Cache, error) {
		called["redis"] = true
		require.Equal(t, "127", addr)
		require.Equal(t, "pw", pwd)
		require.Equal(t, 1, db)
		return &cache.FakeCache{CloseFn: func() error { called["redisClose"] = true; return nil }}, nil
	}
	runMigrationsFn = func(url string) error { called["migrate"] = true; return nil }
	startServer = func(e *echo.Echo, addr string) error { called["start"] = true; return nil }

	t.Setenv("DATABASE_URL", "db")
	t.Setenv("REDIS_ADDR", "127")
	t.Setenv("REDIS_DB", "1")
	t.Setenv("REDIS_PASSWORD", "pw")

	require.NoError(t, run())
	require.True(t, called["pgx"])
	require.True(t, called["redis"])
	require.True(t, called["migrate"])
	require.True(t, called["start"])
	require.True(t, called["dbClose"])
	require.True(t, called["redisClose"])
}

func TestRunErrors(t *testing.T) {
	t.Cleanup(restoreGlobals)
	t.Setenv("DATABASE_URL", "")
	require.Error(t, run())
	t.Setenv("DATABASE_URL", "db")
	t.Setenv("REDIS_ADDR", "")
	require.Error(t, run())
	t.Setenv("REDIS_ADDR", "addr")
	t.Setenv("REDIS_DB", "")
	require.Error(t, run())

	t.Setenv("REDIS_DB", "bad")
	require.Error(t, run())
	t.Setenv("REDIS_DB", "0")
	t.Setenv("REDIS_PASSWORD", "")
	require.Error(t, run())

	t.Setenv("REDIS_PASSWORD", "pw")
	newPgxPool = func(context.Context, string) (database.DB, error) { return nil, errors.New("db") }
	require.Error(t, run())

	newPgxPool = func(context.Context, string) (database.DB, error) { return &database.FakeDB{}, nil }
	newRedisClient = func(string, string, int) (cache.Cache, error) { return nil, errors.New("redis") }
	require.Error(t, run())

	newRedisClient = func(string, string, int) (cache.Cache, error) { return &cache.FakeCache{}, nil }
	runMigrationsFn = func(string) error { return errors.New("migrate") }
	require.Error(t, run())

	runMigrationsFn = func(string) error { return nil }
	startServer = func(*echo.Echo, string) error { return errors.New("start") }
	require.Error(t, run())
}

func TestMainFunction(t *testing.T) {
	t.Cleanup(restoreGlobals)
	startServer = func(*echo.Echo, string) error { return nil }
	newPgxPool = func(context.Context, string) (database.DB, error) { return &database.FakeDB{}, nil }
	newRedisClient = func(string, string, int) (cache.Cache, error) { return &cache.FakeCache{}, nil }
	runMigrationsFn = func(string) error { return nil }
	t.Setenv("DATABASE_URL", "d")
	t.Setenv("REDIS_ADDR", "a")
	t.Setenv("REDIS_DB", "0")
	t.Setenv("REDIS_PASSWORD", "p")
	main()
}

func TestMainExit(t *testing.T) {
	t.Cleanup(restoreGlobals)
	exitCode := 0
	exitFunc = func(code int) { exitCode = code }
	newPgxPool = func(context.Context, string) (database.DB, error) { return nil, errors.New("fail") }
	t.Setenv("DATABASE_URL", "d")
	t.Setenv("REDIS_ADDR", "a")
	t.Setenv("REDIS_DB", "0")
	t.Setenv("REDIS_PASSWORD", "p")
	main()
	require.Equal(t, 1, exitCode)
}
