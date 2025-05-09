// File: internal/handler/users/create_user.go
package users

import (
	"net/http"
	"strings"
	"time"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// CreateUserRequest 定義建立新使用者的請求格式 (form data)
// swagger:model CreateUserRequest
type CreateUserRequest struct {
	// 使用者姓名
	// required: true
	Name string `form:"name" validate:"required" example:"Alice"`

	// 使用者 Email (會自動轉為小寫)
	// required: true
	Email string `form:"email" validate:"required,email" example:"alice@example.com"`

	// 使用者密碼
	// required: true
	Password string `form:"password" validate:"required" example:"Secret123!"`

	// 是否為管理員
	// required: true
	IsAdmin bool `form:"is_admin" validate:"required" example:"false"`
}

// UserResponse 定義回傳的使用者資訊
// swagger:model UserResponse
type UserResponse struct {
	// 使用者 ID
	ID int `json:"id" example:"1"`

	// 使用者姓名
	Name string `json:"name" example:"Alice"`

	// 使用者 Email
	Email string `json:"email" example:"alice@example.com"`

	// 建立時間 (RFC3339 格式)
	CreatedAt time.Time `json:"created_at" example:"2025-05-01T15:04:05Z07:00"`

	// 是否為管理員
	IsAdmin bool `json:"is_admin" example:"false"`
}

// CreateUserHandler 建立新使用者
// @Summary     Create a new user
// @Description 接收使用者表單資料並建立新帳號 (Email 會自動轉小寫)
// @Tags        users
// @Accept      application/x-www-form-urlencoded
// @Produce     json
// @Param       name     formData string true  "使用者姓名"
// @Param       email    formData string true  "使用者 Email (lowercase)"
// @Param       password formData string true  "使用者密碼"
// @Param       is_admin formData boolean true  "是否為管理員"
// @Success     201      {object} UserResponse
// @Failure     400      {object} dto.HTTPError
// @Failure     500      {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Router      /users [post]
func CreateUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateUserRequest
		// Bind 根據 Content-Type 自動綁定 form data
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid form data"})
		}
		// 統一用 validator 驗證所有欄位
		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: err.Error()})
		}

		// Email 轉為小寫以確保一致性
		req.Email = strings.ToLower(req.Email)

		// 密碼哈希
		hash, err := service.HashPassword(req.Password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: "failed to hash password"})
		}

		// 建立使用者模型
		user := &model.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: &hash,
			IsAdmin:      req.IsAdmin,
		}

		// 執行建立操作
		created, err := repository.CreateUser(c.Request().Context(), pool, user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// 組裝並回傳結果
		resp := UserResponse{
			ID:        created.ID,
			Name:      created.Name,
			Email:     created.Email,
			CreatedAt: created.CreatedAt,
			IsAdmin:   created.IsAdmin,
		}
		return c.JSON(http.StatusCreated, resp)
	}
}
