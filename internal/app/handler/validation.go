package handler

import (
	"encoding/json"
	"net/http"
	"toptal/internal/pkg/validator"
)

type ErrorResponse struct {
	Error interface{} `json:"error"`
}

func ValidationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error: "Internal Server Error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	}
}

func WriteValidationError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	var response ErrorResponse
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		response.Error = validationErrors
	} else {
		response.Error = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}
