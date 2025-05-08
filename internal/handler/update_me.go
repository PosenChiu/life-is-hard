// File: internal/handler/update_me.go
package handler

import (
	"net/http"
	"net/mail"
	"strings"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// UpdateMeRequest 定義更新當前使用者資料的請求格式 (form data)
// swagger:model UpdateMeRequest
type UpdateMeRequest struct {
	// 使用者姓名
	// required: true
	Name string `form:"name" validate:"required" example:"Alice"`

	// 使用者 Email (會自動轉為小寫)
	// required: true
	Email string `form:"email" validate:"required,email" example:"alice@example.com"`
}

// UpdateMeHandler 更新當前使用者資料
// @Summary     Update current user info
// @Description 使用 JWT 更新當前使用者姓名和 Email
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       name  formData string true "使用者姓名"
// @Param       email formData string true "使用者 Email (lowercase)"
// @Success     204   "No Content"
// @Failure     400   {object} dto.HTTPError
// @Failure     401   {object} dto.HTTPError
// @Failure     500   {object} dto.HTTPError
// @Security    OAuth2Password[default]
// @Router      /users/me [put]
func UpdateMeHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Bind & Validate
		var req UpdateMeRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid form data"})
		}
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		// Email lowercase & format check
		req.Email = strings.ToLower(req.Email)
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid email format"})
		}

		// Get current user from JWT claims
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}

		// Build model and update
		user := &model.User{
			ID:    claims.ID,
			Name:  req.Name,
			Email: req.Email,
		}
		if err := repository.UpdateUser(c.Request().Context(), pool, user); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}
