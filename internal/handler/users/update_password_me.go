// File: internal/handler/users/update_password_me.go
package users

import (
	"net/http"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// UpdatePasswordMeRequest 定義更新當前使用者密碼的請求格式 (form data)
// swagger:model UpdatePasswordMeRequest
type UpdatePasswordMeRequest struct {
	// 當前密碼
	// required: true
	OldPassword string `form:"old_password" validate:"required" example:"OldSecret123!"`

	// 新密碼
	// required: true
	NewPassword string `form:"new_password" validate:"required" example:"NewSecret456!"`
}

// UpdatePasswordMeHandler 更新當前使用者密碼
// @Summary     Update own password
// @Description 驗證舊密碼並更新為新密碼
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       old_password formData string true "當前密碼"
// @Param       new_password formData string true "新密碼"
// @Success     204      "No Content"
// @Failure     400      {object} dto.HTTPError
// @Failure     401      {object} dto.HTTPError
// @Failure     500      {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Router      /users/me/password [patch]
func UpdatePasswordMeHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req UpdatePasswordMeRequest
		// Bind & Validate
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		// Get current user ID from JWT claims
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}

		// Fetch user from DB
		user, err := repository.GetUserByID(c.Request().Context(), pool, claims.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// Authenticate old password
		if _, err := service.AuthenticateUser(c.Request().Context(), *user, req.OldPassword); err != nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid current password"})
		}

		// Hash new password
		hash, err := service.HashPassword(req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: "failed to hash new password"})
		}

		// Update password
		if err := repository.UpdateUserPassword(c.Request().Context(), pool, claims.ID, hash); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}
