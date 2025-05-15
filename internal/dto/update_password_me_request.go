// File: internal/dto/update_password_me_request.go
package dto

// swagger:model dto.UpdatePasswordMeRequest
type UpdatePasswordMeRequest struct {
	OldPassword string `form:"old_password" validate:"required" example:"OldSecret123!"`
	NewPassword string `form:"new_password" validate:"required" example:"NewSecret456!"`
}
