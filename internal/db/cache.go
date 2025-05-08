// File: internal/db/cache.go
package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 建立並回傳一個 Redis 客戶端實例
// addr: Redis 伺服器位址（如 "localhost:6379"）
// password: Redis 密碼（若無密碼可留空）
// db: 選擇的 Redis 資料庫索引（預設為 0）
// 本函式會在 5 秒內完成 Ping 測試，若連線失敗則返回錯誤。
func NewRedisClient(addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 使用短超時檢查連線
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

// CloseRedisClient 關閉 Redis 客戶端連線
func CloseRedisClient(rdb *redis.Client) error {
	return rdb.Close()
}
