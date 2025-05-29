package api

// swagger:model api.UpdateUserRequest
type UpdateUserRequest struct {
	Name  string `form:"name" validate:"required" example:"Alice"`
	Email string `form:"email" validate:"required,email" example:"alice@example.com"`
}
