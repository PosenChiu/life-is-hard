// File: internal/dto/update_user_o_auth_client_request.go
package dto

// swagger:model dto.UpdateUserOAuthClientRequest
type UpdateUserOAuthClientRequest struct {
	ClientSecret string   `json:"client_secret" validate:"required" example:"new-secret"`
	GrantTypes   []string `json:"grant_types" validate:"required" example:"password,client_credentials,refresh_token"`
}
