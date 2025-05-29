package api

// swagger:model api.HTTPError
type ErrorResponse struct {
	Message string `json:"message"`
}
