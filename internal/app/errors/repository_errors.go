package errors

import (
	"errors"
	"fmt"
)

// Общие ошибки репозиториев
var (
	// ErrNotFound возвращается, когда сущность не найдена
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists возвращается, когда сущность уже существует
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrInvalidData возвращается при невалидных данных
	ErrInvalidData = errors.New("invalid data")

	// ErrDatabase возвращается при ошибке базы данных
	ErrDatabase = errors.New("database error")
)

// Ошибки книг
var (
	// ErrBookNotFound возвращается, когда книга не найдена
	ErrBookNotFound = fmt.Errorf("book %w", ErrNotFound)

	// ErrBookOutOfStock возвращается, когда книга не в наличии
	ErrBookOutOfStock = errors.New("book is out of stock")
)

// Ошибки категорий
var (
	// ErrCategoryNotFound возвращается, когда категория не найдена
	ErrCategoryNotFound = fmt.Errorf("category %w", ErrNotFound)

	// ErrCategoryAlreadyExists возвращается, когда категория уже существует
	ErrCategoryAlreadyExists = fmt.Errorf("category %w", ErrAlreadyExists)
)

// Ошибки пользователей
var (
	// ErrUserNotFound возвращается, когда пользователь не найден
	ErrUserNotFound = fmt.Errorf("user %w", ErrNotFound)

	// ErrUserAlreadyExists возвращается, когда пользователь уже существует
	ErrUserAlreadyExists = fmt.Errorf("user %w", ErrAlreadyExists)

	// ErrInvalidCredentials возвращается при неверных учетных данных
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Ошибки корзины
var (
	// ErrCartEmpty возвращается, когда корзина пуста
	ErrCartEmpty = errors.New("cart is empty")

	// ErrBookAlreadyInCart возвращается, когда книга уже в корзине
	ErrBookAlreadyInCart = errors.New("book already in cart")

	// ErrBookNotInCart возвращается, когда книга не найдена в корзине
	ErrBookNotInCart = errors.New("book not found in cart")
)

// IsNotFound проверяет, является ли ошибка ErrNotFound или его производной
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists проверяет, является ли ошибка ErrAlreadyExists или его производной
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsDatabaseError проверяет, является ли ошибка ErrDatabase
func IsDatabaseError(err error) bool {
	return errors.Is(err, ErrDatabase)
}

// WrapDatabaseError оборачивает ошибку базы данных
func WrapDatabaseError(err error, msg string) error {
	return fmt.Errorf("%s: %w: %v", msg, ErrDatabase, err)
}
