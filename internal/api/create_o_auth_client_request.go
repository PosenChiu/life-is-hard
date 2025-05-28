// File: internal/api/create_o_auth_client_request.go
package api

// swagger:model api.CreateOAuthClientRequest
type CreateOAuthClientRequest struct {
	ClientID     string   `json:"client_id" validate:"required" example:"my-client"`
	ClientSecret string   `json:"client_secret" validate:"required" example:"secret"`
	GrantTypes   []string `json:"grant_types" validate:"required" example:"password,client_credentials,refresh_token"`
}
