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

type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var errors []string
	for _, err := range v {
		errors = append(errors, fmt.Sprintf("%s: %s", err.Field, err.Error))
	}
	return strings.Join(errors, "; ")
}

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

func getErrorMsg(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "field is required"
	case "min":
		return fmt.Sprintf("value must be no less than %s", err.Param())
	case "max":
		return fmt.Sprintf("value should not be greater than %s", err.Param())
	case "email":
		return "invalid email format"
	case "len":
		return fmt.Sprintf("length must be equal %s", err.Param())
	default:
		return fmt.Sprintf("does not comply with the rule: %s", err.Tag())
	}
}
