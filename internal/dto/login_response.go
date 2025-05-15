// File: internal/dto/login_response.go
package dto

import "time"

// swagger:model dto.LoginResponse
type LoginResponse struct {
	AccessToken string    `json:"access_token" example:"eyJhbGciOi..."`
	ExpiresAt   time.Time `json:"expires_at" example:"2025-05-09T15:04:05Z07:00"`
}
