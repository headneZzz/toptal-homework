package handler

import (
	"encoding/json"
	"net/http"
	"toptal/internal/app/handler/model"
	"toptal/internal/pkg/validator"
)

// @Summary User login
// @Description Authenticate user and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.UserRequest true "Login credentials"
// @Success 200 {object} map[string]string "Returns JWT token"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /login [post]
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var request model.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteValidationError(w, err)
		return
	}

	if err := validator.Validate(request); err != nil {
		WriteValidationError(w, err)
		return
	}

	token, err := s.authService.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// @Summary Register new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.UserRequest true "Registration details"
// @Success 201 {object} map[string]string "User created successfully"
// @Failure 400 {string} string "Bad Request"
// @Failure 409 {string} string "Username already exists"
// @Failure 500 {string} string "Internal Server Error"
// @Router /register [post]
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var request model.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteValidationError(w, err)
		return
	}

	if err := validator.Validate(request); err != nil {
		WriteValidationError(w, err)
		return
	}

	if err := s.authService.Register(r.Context(), request.Username, request.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "user created"})
}
