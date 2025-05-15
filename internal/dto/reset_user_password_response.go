// File: internal/dto/reset_user_password_response.go
package dto

// swagger:model dto.ResetUserPasswordResponse
type ResetUserPasswordResponse struct {
	NewPassword string `json:"new_password" example:"Abc123!@#Xyz"`
}
