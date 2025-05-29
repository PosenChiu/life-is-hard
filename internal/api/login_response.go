package api

// swagger:model api.LoginResponse
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOi..."`
}
