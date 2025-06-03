package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type fakeRedis struct {
	FakeCache
	pingErr error
}

func (f *fakeRedis) Ping(ctx context.Context) *redis.StatusCmd {
	return redis.NewStatusResult("PONG", f.pingErr)
}

func TestNewRedisClient(t *testing.T) {
	orig := redisNewClient
	defer func() { redisNewClient = orig }()

	// ping failure
	redisNewClient = func(opt *redis.Options) redisClient { return &fakeRedis{pingErr: errors.New("bad")} }
	c, err := NewRedisClient("addr", "", 0)
	require.Error(t, err)
	require.Nil(t, c)

	// success
	r := &fakeRedis{}
	redisNewClient = func(opt *redis.Options) redisClient { return r }
	c, err = NewRedisClient("addr", "", 0)
	require.NoError(t, err)
	require.Equal(t, r, c)
}
