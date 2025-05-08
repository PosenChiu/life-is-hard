// File: internal/handler/reset_user_password.go
package handler

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// ResetUserPasswordResponse 定義回傳的重置密碼
// swagger:model ResetUserPasswordResponse
type ResetUserPasswordResponse struct {
	// 新的隨機密碼
	NewPassword string `json:"new_password" example:"Abc123!@#Xyz"`
}

// ResetUserPasswordHandler 重置指定使用者密碼並回傳新的隨機密碼
// @Summary     Reset user password
// @Description 由管理員重置特定使用者的密碼，並回傳新的隨機密碼
// @Tags        users
// @Produce     json
// @Param       id   path      int  true  "使用者 ID"
// @Success     200  {object}  ResetUserPasswordResponse
// @Failure     400  {object}  dto.HTTPError
// @Failure     500  {object}  dto.HTTPError
// @Security    OAuth2Password[default]
// @Router      /users/{id}/reset_password [post]
func ResetUserPasswordHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析 path 參數
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid user ID"})
		}

		// 產生新的隨機密碼
		newPwd, err := generateRandomPassword(12)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: "failed to generate password"})
		}

		// 哈希新密碼
		hash, err := service.HashPassword(newPwd)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: "failed to hash password"})
		}

		// 更新資料庫
		if err := repository.UpdateUserPassword(c.Request().Context(), pool, id, hash); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// 回傳新密碼
		resp := ResetUserPasswordResponse{NewPassword: newPwd}
		return c.JSON(http.StatusOK, resp)
	}
}

// generateRandomPassword 產生指定長度的隨機密碼，包含大寫、小寫、數字與符號
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" +
		"!@#$%^&*()-_=+[]{}<>?"
	pwd := make([]byte, length)
	for i := 0; i < length; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		pwd[i] = charset[nBig.Int64()]
	}
	return string(pwd), nil
}
