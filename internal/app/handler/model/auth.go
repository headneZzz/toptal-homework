package model

type AuthRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=50"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}
