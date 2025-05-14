// File: internal/handler/ping.go
package handler

import (
	"net/http"
	"time"

	"life-is-hard/internal/dto"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// PingResponse 健康檢查回應模型
// swagger:model PingResponse
type PingResponse struct {
	// 回應訊息
	Message string `json:"message" example:"pong"`
}

// PingHandler 健康檢查（需通過認證），並於 Redis 設定示例鍵值
// @Summary     Health Check
// @Description 回傳 pong，並檢查資料庫連線是否正常，同時在 Redis 設置一個示例鍵值
// @Tags        ping
// @Accept      json
// @Produce     json
// @Success     200 {object} PingResponse
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Router      /ping [get]
func PingHandler(db *pgxpool.Pool, rdb *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		// 檢查資料庫連線
		if err := db.Ping(ctx); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: "database unhealthy"})
		}

		// 在 Redis 設置示例鍵值
		err := rdb.Set(ctx, "ping:timestamp", time.Now().Format(time.RFC3339), time.Minute).Err()
		if err != nil {
			// 若 Redis 設定失敗，不影響主要流程，但記錄日誌
			c.Logger().Error("Redis SET failed: ", err)
		}

		// 回傳正常訊息
		return c.JSON(http.StatusOK, PingResponse{Message: "pong"})
	}
}
