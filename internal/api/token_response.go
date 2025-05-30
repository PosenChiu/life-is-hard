package api

// swagger:model api.TokenResponse
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int    `json:"expires_in" example:"86400"`
	RefreshToken string `json:"refresh_token,omitempty" example:"..."`
}
