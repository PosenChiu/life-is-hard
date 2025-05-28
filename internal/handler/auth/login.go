// File: internal/handler/auth/login.go
package auth

import (
	"fmt"
	"net/http"
	"time"

	"life-is-hard/internal/api"
	"life-is-hard/internal/service"
	"life-is-hard/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// LoginHandler 使用 Username/Password 驗證並回傳 JWT
// @Summary     登入使用者
// @Description 使用 Username 與 Password 進行驗證，回傳存取令牌與到期時間
// @Tags        auth
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       username formData string true "使用者名稱"
// @Param       password formData string true "使用者密碼"
// @Success     200      {object} api.LoginResponse
// @Failure     400      {object} api.ErrorResponse
// @Failure     401      {object} api.ErrorResponse
// @Failure     500      {object} api.ErrorResponse
// @Router      /auth/login [post]
func LoginHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req api.LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: fmt.Sprintf("無效的表單資料: %v", err)})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, api.ErrorResponse{Message: err.Error()})
		}

		user, err := store.GetUserByName(c.Request().Context(), pool, req.Username)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid credentials"})
		}
		if err := service.AuthenticateUser(c.Request().Context(), *user, req.Password); err != nil {
			return c.JSON(http.StatusUnauthorized, api.ErrorResponse{Message: "invalid credentials"})
		}

		token, err := service.IssueAccessToken(*user, 24*time.Hour)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{Message: fmt.Sprintf("failed to issue token: %v", err)})
		}

		return c.JSON(http.StatusOK, api.LoginResponse{AccessToken: token})
	}
}
