package domain

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInternalServer = errors.New("internal server error")
	ErrUnauthorized   = errors.New("unauthorized")
)
