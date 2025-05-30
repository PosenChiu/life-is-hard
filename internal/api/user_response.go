package api

import "time"

// swagger:model api.UserResponse
type UserResponse struct {
	ID        int       `json:"id" example:"1"`
	Name      string    `json:"name" example:"Alice"`
	Email     string    `json:"email" example:"alice@example.com"`
	IsAdmin   bool      `json:"is_admin" example:"false"`
	CreatedAt time.Time `json:"created_at" example:"2025-05-01T15:04:05Z07:00"`
}
