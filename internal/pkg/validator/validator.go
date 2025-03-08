package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ValidationErrors представляет список ошибок валидации
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var errors []string
	for _, err := range v {
		errors = append(errors, fmt.Sprintf("%s: %s", err.Field, err.Error))
	}
	return strings.Join(errors, "; ")
}

// Validate проверяет структуру на соответствие тегам валидации
func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors ValidationErrors
	for _, err := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, ValidationError{
			Field: err.Field(),
			Error: getErrorMsg(err),
		})
	}
	return validationErrors
}

// getErrorMsg возвращает понятное сообщение об ошибке на основе тега валидации
func getErrorMsg(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "это поле обязательно"
	case "min":
		return fmt.Sprintf("значение должно быть не меньше %s", err.Param())
	case "max":
		return fmt.Sprintf("значение должно быть не больше %s", err.Param())
	case "email":
		return "неверный формат email"
	case "len":
		return fmt.Sprintf("длина должна быть равна %s", err.Param())
	default:
		return fmt.Sprintf("не соответствует правилу: %s", err.Tag())
	}
}
