// File: internal/dto/update_my_password_request.go
package dto

// swagger:model dto.UpdateMyPasswordRequest
type UpdateMyPasswordRequest struct {
	OldPassword string `form:"old_password" validate:"required" example:"OldSecret123!"`
	NewPassword string `form:"new_password" validate:"required" example:"NewSecret456!"`
}
