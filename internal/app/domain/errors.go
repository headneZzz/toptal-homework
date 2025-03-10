package domain

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInternalServer  = errors.New("internal server error")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidCategory = errors.New("invalid category")
)
