package repository

import (
	"context"
	"database/sql"
	"toptal/internal/pkg/pg"

	"github.com/jmoiron/sqlx"

	"toptal/internal/app/domain"
	appErrors "toptal/internal/app/errors"
)

// SQL запросы для книг
const (
	sqlCreateBook = `
		INSERT INTO books (title, author, year, price, stock, category_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	sqlGetBookById = `SELECT * FROM books WHERE id = $1`
	sqlUpdateBook  = `
		UPDATE books
		SET title = $2, author = $3, year = $4, price = $5, category_id = $6
		WHERE id = $1
	`
	sqlDeleteBook           = `DELETE FROM books WHERE id = $1`
	sqlGetBooks             = `SELECT * FROM books WHERE stock > 0`
	sqlGetBooksByCategories = `SELECT * FROM books WHERE stock > 0 AND category_id IN (?)`
)

// BookRepository представляет репозиторий для работы с книгами
type BookRepository struct {
	DB *pg.DB
}

// NewBookRepository создает новый репозиторий для работы с книгами
func NewBookRepository(db *pg.DB) *BookRepository {
	return &BookRepository{
		DB: db,
	}
}

// Create создает новую книгу
func (r *BookRepository) Create(ctx context.Context, book domain.Book) error {
	var id int
	err := r.DB.QueryRow(ctx, "create_book", sqlCreateBook,
		book.Title, book.Author, book.Year, book.Price, book.Stock, book.CategoryId,
	).Scan(&id)

	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to create book")
	}

	return nil
}

// GetById получает книгу по идентификатору
func (r *BookRepository) GetById(ctx context.Context, id int) (domain.Book, error) {
	var book domain.Book
	err := r.DB.Get(ctx, "get_book_by_id", &book, sqlGetBookById, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Book{}, appErrors.ErrBookNotFound
		}
		return domain.Book{}, appErrors.WrapDatabaseError(err, "failed to get book")
	}

	return book, nil
}

// Update обновляет информацию о книге
func (r *BookRepository) Update(ctx context.Context, book domain.Book) error {
	result, err := r.DB.Exec(ctx, "update_book", sqlUpdateBook,
		book.Id, book.Title, book.Author, book.Year, book.Price, book.CategoryId,
	)

	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to update book")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to get affected rows")
	}

	if affected == 0 {
		return appErrors.ErrBookNotFound
	}

	return nil
}

func (r *BookRepository) Delete(ctx context.Context, id int) error {
	result, err := r.DB.Exec(ctx, "delete_book", sqlDeleteBook, id)

	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to delete book")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to get affected rows")
	}

	if affected == 0 {
		return appErrors.ErrBookNotFound
	}

	return nil
}

// GetAll получает все книги
func (r *BookRepository) GetAll(ctx context.Context) ([]domain.Book, error) {
	var books []domain.Book
	err := r.DB.Select(ctx, "get_all_books", &books, sqlGetBooks)

	if err != nil {
		return nil, appErrors.WrapDatabaseError(err, "failed to get books")
	}

	return books, nil
}

// GetByCategories получает книги по нескольким категориям
func (r *BookRepository) GetByCategories(ctx context.Context, categoryIds []int) ([]domain.Book, error) {
	var books []domain.Book

	// Если категории не указаны, возвращаем все книги
	if len(categoryIds) == 0 {
		return r.GetAll(ctx)
	}

	query, args, err := sqlx.In(sqlGetBooksByCategories, categoryIds)
	if err != nil {
		return nil, appErrors.WrapDatabaseError(err, "failed to build IN query")
	}

	// Заменяем ? на $1, $2, ... для Postgres
	query = r.DB.DB.Rebind(query)

	err = r.DB.Select(ctx, "get_books_by_categories", &books, query, args...)
	if err != nil {
		return nil, appErrors.WrapDatabaseError(err, "failed to get books by categories")
	}

	return books, nil
}
