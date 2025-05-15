// File: internal/dto/create_o_auth_client_request.go
package dto

// swagger:model dto.CreateOAuthClientRequest
type CreateOAuthClientRequest struct {
	ClientID     string   `json:"client_id" validate:"required" example:"my-client"`
	ClientSecret string   `json:"client_secret" validate:"required" example:"secret"`
	OwnerID      int      `json:"owner_id" example:"1"`
	GrantTypes   []string `json:"grant_types" validate:"required" example:"password,client_credentials,refresh_token"`
}
