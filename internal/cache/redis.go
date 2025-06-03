package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// redisClient 抽象化 *redis.Client 方便測試時替換
type redisClient interface {
	Cache
	Ping(ctx context.Context) *redis.StatusCmd
}

var redisNewClient = func(opt *redis.Options) redisClient { return redis.NewClient(opt) }

// NewRedisClient 建立並回傳 *redis.Client，直接實作 Cache
// addr: Redis 位址；password: 密碼，可空；db: 資料庫編號
func NewRedisClient(addr string, password string, db int) (Cache, error) {
	client := redisNewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
