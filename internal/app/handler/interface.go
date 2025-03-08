package handler

import (
	"context"
	"toptal/internal/app/domain"
)

type BookService interface {
	GetBookById(ctx context.Context, id int) (domain.Book, error)
	GetAvailableBooks(ctx context.Context, categoryIds []int) ([]domain.Book, error)
	CreateBook(ctx context.Context, book domain.Book) error
	UpdateBook(ctx context.Context, book domain.Book) error
	DeleteBook(ctx context.Context, id int) error
}

type CategoryService interface {
	GetCategoryById(ctx context.Context, id int) (domain.Category, error)
	GetCategories(ctx context.Context) ([]domain.Category, error)
	CreateCategory(ctx context.Context, book domain.Category) error
	UpdateCategory(ctx context.Context, book domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type UserService interface {
	Login(ctx context.Context, username string, password string) (string, error)
	CreateUser(ctx context.Context, username string, password string) error
}

type CartService interface {
	GetCart(ctx context.Context, userId int) ([]domain.Book, error)
	AddToCart(ctx context.Context, userId int, bookId int) error
	RemoveFromCart(ctx context.Context, userId int, bookId int) error
	Purchase(ctx context.Context, userId int) error
}

type HealthService interface {
	CheckDatabase(ctx context.Context) error
}
