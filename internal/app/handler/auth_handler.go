package handler

import (
	"encoding/json"
	"net/http"
	"toptal/internal/app/domain"
	"toptal/internal/app/handler/model"
	"toptal/internal/pkg/validator"
)

// @Summary User login
// @Description Authenticate user and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.AuthRequest true "Login credentials"
// @Success 200 {object} model.LoginResponse "Returns JWT token"
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 401 {object} model.ProblemDetail "Unauthorized"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Router /login [post]
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var request model.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		model.InvalidRequest(w, err.Error(), r.URL.Path)
		return
	}

	if err := validator.Validate(request); err != nil {
		model.ValidationError(w, err.Error(), r.URL.Path)
		return
	}

	token, err := s.authService.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		model.Unauthorized(w, domain.ErrUnauthorized.Error(), err.Error())
		return
	}

	response := model.LoginResponse{Token: token}
	writeResponseOK(w, response)
}

// @Summary Register new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.AuthRequest true "Registration details"
// @Success 201 {object} model.RegisterResponse "User created successfully"
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 409 {object} model.ProblemDetail "Username already exists"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Router /register [post]
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var request model.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		model.InvalidRequest(w, err.Error(), r.URL.Path)
		return
	}

	if err := validator.Validate(request); err != nil {
		model.ValidationError(w, err.Error(), r.URL.Path)
		return
	}

	if err := s.authService.Register(r.Context(), request.Username, request.Password); err != nil {
		model.InternalServerError(w, r.URL.Path)
		return
	}

	response := model.RegisterResponse{Message: "User created successfully"}
	writeResponseCreated(w, response)
}
