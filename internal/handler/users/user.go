// File: internal/handler/users/users.go
package users

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"life-is-hard/internal/dto"
	"life-is-hard/internal/model"
	"life-is-hard/internal/repository"
	"life-is-hard/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

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
// @Success     201      {object} dto.UserResponse
// @Failure     400      {object} dto.HTTPError
// @Failure     500      {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users [post]
func CreateUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.CreateUserRequest
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
		resp := dto.UserResponse{
			ID:        created.ID,
			Name:      created.Name,
			Email:     created.Email,
			CreatedAt: created.CreatedAt,
			IsAdmin:   created.IsAdmin,
		}
		return c.JSON(http.StatusCreated, resp)
	}
}

// DeleteMeHandler 刪除當前使用者帳號
// @Summary     Delete current user
// @Description 使用 JWT Token 刪除當前使用者帳號
// @Tags        users
// @Produce     json
// @Success     204
// @Failure     401 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [delete]
func DeleteMeHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 從 context 中取出 JWT claims
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}

		// 刪除當前使用者
		if err := repository.DeleteUser(c.Request().Context(), pool, claims.UserID); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// 回傳 No Content
		return c.NoContent(http.StatusNoContent)
	}
}

// DeleteUserHandler 刪除指定 ID 的使用者
// @Summary     Delete a user by ID
// @Description 根據使用者 ID 刪除使用者帳號
// @Tags        users
// @Param       id   path      int  true  "使用者 ID"
// @Success     204  "No Content"
// @Failure     400  {object}  dto.HTTPError  "參數錯誤"
// @Failure     500  {object}  dto.HTTPError  "伺服器錯誤"
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{id} [delete]
func DeleteUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析 path 參數
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid user ID"})
		}

		// 執行刪除操作
		if err := repository.DeleteUser(c.Request().Context(), pool, id); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// 回傳 204 No Content
		return c.NoContent(http.StatusNoContent)
	}
}

// GetMeHandler 取得當前使用者資訊
// @Summary     Get current user info
// @Description 透過 JWT Token 取得當前使用者詳細資訊
// @Tags        users
// @Produce     json
// @Success     200 {object} dto.UserResponse
// @Failure     401 {object} dto.HTTPError
// @Failure     500 {object} dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [get]
func GetMeHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 從 context 取得 JWT claims
		claimsRaw := c.Get("user")
		claims, ok := claimsRaw.(*service.CustomClaims)
		if !ok || claimsRaw == nil {
			return c.JSON(http.StatusUnauthorized, dto.HTTPError{Message: "invalid or missing token"})
		}

		// 根據 claims.ID 查詢使用者
		user, err := repository.GetUserByID(c.Request().Context(), pool, claims.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		// 組裝回應
		resp := dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		}
		return c.JSON(http.StatusOK, resp)
	}
}

// GetUserHandler 透過使用者 ID 取得使用者資訊
// @Summary     Get a user by ID
// @Description 透過 ID 查詢並回傳使用者詳細資料
// @Tags        users
// @Produce     json
// @Param       id   path      int  true  "使用者 ID"
// @Success     200  {object}  dto.UserResponse
// @Failure     400  {object}  dto.HTTPError  "參數錯誤"
// @Failure     404  {object}  dto.HTTPError  "使用者不存在"
// @Failure     500  {object}  dto.HTTPError  "伺服器錯誤"
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{id} [get]
func GetUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析 path 參數
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid user ID"})
		}

		// 從資料庫讀取使用者資料
		user, err := repository.GetUserByID(c.Request().Context(), pool, id)
		if err != nil {
			return c.JSON(http.StatusNotFound, dto.HTTPError{Message: "user not found"})
		}

		// 組裝回傳用結構
		resp := dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		}
		return c.JSON(http.StatusOK, resp)
	}
}

// ResetUserPasswordHandler 重置指定使用者密碼並回傳新的隨機密碼
// @Summary     Reset user password
// @Description 由管理員重置特定使用者的密碼，並回傳新的隨機密碼
// @Tags        users
// @Produce     json
// @Param       id   path      int  true  "使用者 ID"
// @Success     200  {object}  dto.ResetUserPasswordResponse
// @Failure     400  {object}  dto.HTTPError
// @Failure     500  {object}  dto.HTTPError
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
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
		resp := dto.ResetUserPasswordResponse{NewPassword: newPwd}
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
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me [put]
func UpdateMeHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Bind & Validate
		var req dto.UpdateMeRequest
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
			ID:    claims.UserID,
			Name:  req.Name,
			Email: req.Email,
		}
		if err := repository.UpdateUser(c.Request().Context(), pool, user); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
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
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/me/password [patch]
func UpdatePasswordMeHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req dto.UpdatePasswordMeRequest
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
		user, err := repository.GetUserByID(c.Request().Context(), pool, claims.UserID)
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
		if err := repository.UpdateUserPassword(c.Request().Context(), pool, claims.UserID, hash); err != nil {
			return c.JSON(http.StatusInternalServerError, dto.HTTPError{Message: err.Error()})
		}

		return c.NoContent(http.StatusNoContent)
	}
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
// @Security    ApiKeyAuth
// @Security    OAuth2Application
// @Security    OAuth2Password
// @Router      /users/{id} [put]
func UpdateUserHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 解析 ID
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.HTTPError{Message: "invalid user ID"})
		}

		var req dto.UpdateUserRequest
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
