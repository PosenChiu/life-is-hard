// File: internal/dto/login_request.go
package dto

// swagger:model dto.LoginRequest
type LoginRequest struct {
	Username string `form:"username" validate:"required" example:"alice"`
	Password string `form:"password" validate:"required" example:"Secret123!"`
}
