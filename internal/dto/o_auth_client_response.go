// File: internal/dto/o_auth_client_response.go
package dto

import "time"

// swagger:model dto.OAuthClientResponse
type OAuthClientResponse struct {
	ID           int       `json:"id" example:"1"`
	ClientID     string    `json:"client_id" example:"my-client"`
	ClientSecret string    `json:"client_secret" example:"secret"`
	OwnerID      int       `json:"owner_id" example:"42"`
	GrantTypes   []string  `json:"grant_types" example:"password,client_credentials"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
