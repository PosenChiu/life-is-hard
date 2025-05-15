// File: internal/dto/update_me_request.go
package dto

// swagger:model dto.UpdateMeRequest
type UpdateMeRequest struct {
	Name  string `form:"name" validate:"required" example:"Alice"`
	Email string `form:"email" validate:"required,email" example:"alice@example.com"`
}
