package domain

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidCategory = errors.New("invalid category")
	ErrBookNotFound    = errors.New("book not found")
	ErrBookOutOfStock  = errors.New("book out of stock")
	ErrBookNotInCart   = errors.New("book not in cart")
	ErrCartEmpty       = errors.New("cart is empty")
)
