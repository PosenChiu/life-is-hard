// File: internal/dto/update_user_request.go
package dto

// swagger:model dto.UpdateUserRequest
type UpdateUserRequest struct {
	Name  string `form:"name" validate:"required" example:"Alice"`
	Email string `form:"email" validate:"required,email" example:"alice@example.com"`
}
