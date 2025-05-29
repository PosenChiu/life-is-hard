package api

// swagger:model api.UpdateOAuthClientRequest
type UpdateOAuthClientRequest struct {
	ClientSecret string   `json:"client_secret" validate:"required" example:"new-secret"`
	GrantTypes   []string `json:"grant_types" validate:"required" example:"password,client_credentials,refresh_token"`
}
