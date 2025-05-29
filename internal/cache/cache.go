package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache 定義快取操作介面
// 提供基礎的 Get、Set、Close 方法
// 用於封裝 Redis 或其他快取實作
// 方便測試時替換 FakeCache 實作
// ttl <= 0 表示不設過期

type Cache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd
	Close() error
}

type FakeCache struct {
	GetFn   func(ctx context.Context, key string) *redis.StringCmd
	SetFn   func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	CloseFn func() error
}

// Get 執行 Fake 設定或 panic
func (f *FakeCache) Get(ctx context.Context, key string) *redis.StringCmd {
	if f.GetFn != nil {
		return f.GetFn(ctx, key)
	}
	panic("unexpected Get")
}

// Set 執行 Fake 設定或 panic
func (f *FakeCache) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	if f.SetFn != nil {
		return f.SetFn(ctx, key, value, expiration)
	}
	panic("unexpected Set")
}

// Close 執行 Fake 設定或 no-op
func (f *FakeCache) Close() error {
	if f.CloseFn != nil {
		return f.CloseFn()
	}
	return nil
}
