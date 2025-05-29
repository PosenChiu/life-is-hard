package api

import "time"

// swagger:model api.OAuthClientResponse
type OAuthClientResponse struct {
	ClientID     string    `json:"client_id" example:"my-client"`
	ClientSecret string    `json:"client_secret" example:"secret"`
	UserID       int       `json:"user_id" example:"42"`
	GrantTypes   []string  `json:"grant_types" example:"password,client_credentials"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
