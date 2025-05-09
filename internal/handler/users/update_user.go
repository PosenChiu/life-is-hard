// File: internal/handler/users/update_user.go
package users

import (
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// UpdateUserRequest 定義更新使用者資料的請求格式 (form data)
// swagger:model UpdateUserRequest
type UpdateUserRequest struct {
	// 使用者姓名
	// required: true
	Name string `form:"name" validate:"required" example:"Alice"`

	// 使用者 Email (會自動轉為小寫)
	// required: true
	Email string `form:"email" validate:"required,email" example:"alice@example.com"`

	// 是否為管理員
	// required: true
	IsAdmin bool `form:"is_admin" validate:"required" example:"false"`
}

// UpdateUserHandler 更新指定使用者資料
// @Summary     Update a user by ID
// @Description 根據使用者 ID 更新使用者姓名、Email 及管理員狀態
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       id       path     int    true  "使用者 ID"
// @Param       name     formData string true  "使用者姓名"
// @Param       email    formData string true  "使用者 Email (lowercase)"
// @Param       is_admin formData boolean true "是否為管理員"
// @Success     204      "No Content"
// @Failure     400      {object} dto.HTTPError
// @Failure     404      {object} dto.HTTPError
// @Failure     500      {object} dto.HTTPError
// @Security    OAuth2Password[default]
// @Router      /users/{id} [put]
func UpdateUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析 ID
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid user ID"})
		}

		var req UpdateUserRequest
		// Bind & Validate
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

		// 构建模型并更新
		user := &model.User{
			ID:      id,
			Name:    req.Name,
			Email:   req.Email,
			IsAdmin: req.IsAdmin,
		}
		if err := repository.UpdateUser(c.Request().Context(), pool, user); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
}
