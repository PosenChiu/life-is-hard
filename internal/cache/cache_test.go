package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestFakeCache(t *testing.T) {
	c := &FakeCache{}
	require.Panics(t, func() { c.Get(context.Background(), "k") })
	require.Panics(t, func() { c.Set(context.Background(), "k", 1, 0) })
	require.NoError(t, c.Close())

	gCalled := false
	sCalled := false
	clCalled := false
	c.GetFn = func(ctx context.Context, key string) *redis.StringCmd {
		gCalled = true
		return redis.NewStringResult("v", nil)
	}
	c.SetFn = func(ctx context.Context, key string, val any, exp time.Duration) *redis.StatusCmd {
		sCalled = true
		return redis.NewStatusResult("OK", nil)
	}
	c.CloseFn = func() error { clCalled = true; return errors.New("close") }

	require.Equal(t, "v", c.Get(context.Background(), "k").Val())
	require.Equal(t, "OK", c.Set(context.Background(), "k", 1, 0).Val())
	require.EqualError(t, c.Close(), "close")
	require.True(t, gCalled)
	require.True(t, sCalled)
	require.True(t, clCalled)
}
