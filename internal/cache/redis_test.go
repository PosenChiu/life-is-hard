package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// stubClient implements redisClient for testing.
type stubClient struct {
	pingErr error
}

func (s *stubClient) Ping(ctx context.Context) *redis.StatusCmd {
	return redis.NewStatusResult("PONG", s.pingErr)
}

func (s *stubClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return redis.NewStringResult("", nil)
}

func (s *stubClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	return redis.NewStatusResult("OK", nil)
}

func (s *stubClient) Close() error { return nil }

func TestNewRedisClient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var opts *redis.Options
		stub := &stubClient{}
		redisNewClient = func(o *redis.Options) redisClient {
			opts = o
			return stub
		}
		defer func() { redisNewClient = func(o *redis.Options) redisClient { return redis.NewClient(o) } }()

		c, err := NewRedisClient("127.0.0.1:6379", "secret", 1)
		require.NoError(t, err)
		require.Equal(t, stub, c)
		require.Equal(t, "127.0.0.1:6379", opts.Addr)
		require.Equal(t, "secret", opts.Password)
		require.Equal(t, 1, opts.DB)
	})

	t.Run("ping fail", func(t *testing.T) {
		redisNewClient = func(o *redis.Options) redisClient {
			return &stubClient{pingErr: errors.New("fail")}
		}
		defer func() { redisNewClient = func(o *redis.Options) redisClient { return redis.NewClient(o) } }()

		c, err := NewRedisClient("addr", "", 0)
		require.Error(t, err)
		require.Nil(t, c)
	})
}
