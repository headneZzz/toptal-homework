package service

import (
	"context"
	"toptal/internal/app/domain"
)

type BookRepository interface {
	Create(ctx context.Context, book domain.Book) error
	GetById(ctx context.Context, id int) (domain.Book, error)
	GetByCategories(ctx context.Context, categoryIds []int) ([]domain.Book, error)
	Update(ctx context.Context, book domain.Book) error
	Delete(ctx context.Context, id int) error
}

type CategoryRepository interface {
	InsertCategory(ctx context.Context, book domain.Category) error
	FindCategoryById(ctx context.Context, id int) (domain.Category, error)
	FindCategories(ctx context.Context) ([]domain.Category, error)
	UpdateCategory(ctx context.Context, category domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type UserRepository interface {
	FindUserByName(ctx context.Context, name string) (domain.User, error)
	FindUserById(ctx context.Context, id int) (domain.User, error)
	CreateUser(ctx context.Context, user domain.User) error
}

type CartRepository interface {
	GetCart(ctx context.Context, userId int) ([]domain.Book, error)
	AddToCart(ctx context.Context, userId int, bookId int) error
	RemoveFromCart(ctx context.Context, userId int, bookId int) error
	Purchase(ctx context.Context, userId int) error
}
