// File: internal/api/error_response.go
package api

// swagger:model api.HTTPError
type ErrorResponse struct {
	Message string `json:"message"`
}
