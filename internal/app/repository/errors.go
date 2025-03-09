package repository

import (
	"errors"
	"fmt"
)

// common
var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
	ErrInvalidData   = errors.New("invalid data")
	ErrDatabase      = errors.New("database error")
)

// book
var (
	ErrBookNotFound   = fmt.Errorf("book %w", ErrNotFound)
	ErrBookOutOfStock = errors.New("book is out of stock")
)

// category
var (
	ErrCategoryNotFound      = fmt.Errorf("category %w", ErrNotFound)
	ErrCategoryAlreadyExists = fmt.Errorf("category %w", ErrAlreadyExists)
)

// user
var (
	ErrUserNotFound       = fmt.Errorf("user %w", ErrNotFound)
	ErrUserAlreadyExists  = fmt.Errorf("user %w", ErrAlreadyExists)
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// cart
var (
	ErrCartEmpty         = errors.New("cart is empty")
	ErrBookAlreadyInCart = errors.New("book already in cart")
	ErrBookNotInCart     = errors.New("book not found in cart")
)

func WrapDatabaseError(err error, msg string) error {
	return fmt.Errorf("%s: %w: %v", msg, ErrDatabase, err)
}
