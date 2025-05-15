// File: internal/dto/user_response.go
package dto

import "time"

// swagger:model dto.UserResponse
type UserResponse struct {
	ID        int       `json:"id" example:"1"`
	Name      string    `json:"name" example:"Alice"`
	Email     string    `json:"email" example:"alice@example.com"`
	CreatedAt time.Time `json:"created_at" example:"2025-05-01T15:04:05Z07:00"`
	IsAdmin   bool      `json:"is_admin" example:"false"`
}
