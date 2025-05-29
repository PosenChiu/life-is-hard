package api

// swagger:model api.CreateUserRequest
type CreateUserRequest struct {
	Name     string `form:"name" validate:"required" example:"Alice"`
	Email    string `form:"email" validate:"required,email" example:"alice@example.com"`
	Password string `form:"password" validate:"required" example:"Secret123!"`
	IsAdmin  bool   `form:"is_admin" validate:"required" example:"false"`
}
