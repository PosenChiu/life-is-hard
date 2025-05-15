// File: internal/dto/update_o_auth_client_request.go
package dto

// swagger:model UpdateOAuthClientRequest
type UpdateOAuthClientRequest struct {
	ClientSecret string   `json:"client_secret" validate:"required" example:"new-secret"`
	OwnerID      int      `json:"owner_id" example:"2"`
	GrantTypes   []string `json:"grant_types" validate:"required" example:"password,client_credentials,refresh_token"`
}
