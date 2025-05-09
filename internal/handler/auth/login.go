// File: internal/handler/auth/login.go
package auth

import (
	"fmt"
	"net/http"
	"time"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// LoginRequest 定義表單登入請求資料
// swagger:model LoginRequest
type LoginRequest struct {
	// 使用者名稱
	// required: true
	Username string `form:"username" validate:"required" example:"alice"`

	// 使用者密碼
	// required: true
	Password string `form:"password" validate:"required" example:"Secret123!"`
}

// LoginResponse 定義回傳的存取令牌與過期時間
// swagger:model LoginResponse
type LoginResponse struct {
	// 存取令牌
	AccessToken string `json:"access_token" example:"eyJhbGciOi..."`
	// 到期時間 (RFC3339 格式)
	ExpiresAt time.Time `json:"expires_at" example:"2025-05-09T15:04:05Z07:00"`
}

// LoginHandler 使用 Username/Password 驗證並回傳 JWT
// @Summary     登入使用者
// @Description 使用 Username 與 Password 進行驗證，回傳存取令牌與到期時間
// @Tags        auth
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       username formData string true "使用者名稱"
// @Param       password formData string true "使用者密碼"
// @Success     200      {object} LoginResponse
// @Failure     400      {object} dto.HTTPError
// @Failure     401      {object} dto.HTTPError
// @Failure     500      {object} dto.HTTPError
// @Router      /auth/login [post]
func LoginHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest
		// 先 Bind
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: fmt.Sprintf("無效的表單資料: %v", err)})
		}
		// 再驗證結構化參數 (go-playground/validator)
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		// 撈使用者資料
		user, err := repository.GetUserByName(c.Request().Context(), pool, req.Username)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid credentials"})
		}

		// 驗證密碼
		authUser, err := service.AuthenticateUser(c.Request().Context(), *user, req.Password)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid credentials"})
		}

		// 發行存取令牌
		token, err := service.IssueAccessToken(*authUser, 24*time.Hour)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: fmt.Sprintf("failed to issue token: %v", err)})
		}

		// 回傳 JWT 及到期時間
		expiresAt := time.Now().Add(24 * time.Hour)
		resp := LoginResponse{
			AccessToken: token,
			ExpiresAt:   expiresAt,
		}
		return c.JSON(http.StatusOK, resp)
	}
}
