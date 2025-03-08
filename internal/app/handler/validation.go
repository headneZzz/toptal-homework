package handler

import (
	"encoding/json"
	"net/http"
	"toptal/internal/pkg/validator"
)

// ErrorResponse представляет структуру ответа с ошибкой
type ErrorResponse struct {
	Error interface{} `json:"error"`
}

// ValidationMiddleware добавляет обработку ошибок валидации
func ValidationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error: "внутренняя ошибка сервера",
				})
			}
		}()

		next.ServeHTTP(w, r)
	}
}

// WriteValidationError записывает ошибку валидации в ответ
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
