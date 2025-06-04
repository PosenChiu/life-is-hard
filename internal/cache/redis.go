package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// redisClient 定義了 NewRedisClient 內部使用的必要方法，便於測試時替換。
type redisClient interface {
	Cache
	Ping(ctx context.Context) *redis.StatusCmd
}

// redisNewClient 用來建立 redis client，測試可覆寫此變數。
var redisNewClient = func(opt *redis.Options) redisClient {
	return redis.NewClient(opt)
}

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
