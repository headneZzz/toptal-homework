package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"toptal/internal/pkg/pg"

	"toptal/internal/app/domain"
)

const (
	sqlCreateBook = `
		INSERT INTO books (title, author, year, price, stock, category_id)
		VALUES ($1, $2, $3, $4, $5, $6)
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

type BookRepository struct {
	db *pg.DB
}

func NewBookRepository(db *pg.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) GetById(ctx context.Context, id int) (domain.Book, error) {
	var book domain.Book
	err := r.db.Get(ctx, "get_book_by_id", &book, sqlGetBookById, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Book{}, ErrBookNotFound
		}
		return domain.Book{}, WrapDatabaseError(err, "failed to get book")
	}

	return book, nil
}

func (r *BookRepository) GetAll(ctx context.Context) ([]domain.Book, error) {
	var books []domain.Book
	err := r.db.Select(ctx, "get_all_books", &books, sqlGetBooks)

	if err != nil {
		return nil, WrapDatabaseError(err, "failed to get books")
	}

	return books, nil
}

// TODO pagination
func (r *BookRepository) GetByCategories(ctx context.Context, categoryIds []int) ([]domain.Book, error) {
	var books []domain.Book
	if len(categoryIds) == 0 {
		return r.GetAll(ctx)
	}

	query, args, err := sqlx.In(sqlGetBooksByCategories, categoryIds)
	if err != nil {
		return nil, WrapDatabaseError(err, "failed to build IN query")
	}
	query = r.db.Rebind(query)

	err = r.db.Select(ctx, "get_books_by_categories", &books, query, args...)
	if err != nil {
		return nil, WrapDatabaseError(err, "failed to get books by categories")
	}

	return books, nil
}

func (r *BookRepository) Create(ctx context.Context, book domain.Book) error {
	_, err := r.db.Exec(ctx, "create_book", sqlCreateBook,
		book.Title, book.Author, book.Year, book.Price, book.Stock, book.CategoryId,
	)

	if err != nil {
		return WrapDatabaseError(err, "failed to create book")
	}

	return nil
}

func (r *BookRepository) Update(ctx context.Context, book domain.Book) error {
	result, err := r.db.Exec(ctx, "update_book", sqlUpdateBook,
		book.Id, book.Title, book.Author, book.Year, book.Price, book.CategoryId,
	)

	if err != nil {
		return WrapDatabaseError(err, "failed to update book")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return WrapDatabaseError(err, "failed to get affected rows")
	}

	if affected == 0 {
		return ErrBookNotFound
	}

	return nil
}

func (r *BookRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx, "delete_book", sqlDeleteBook, id)

	if err != nil {
		return WrapDatabaseError(err, "failed to delete book")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return WrapDatabaseError(err, "failed to get affected rows")
	}

	if affected == 0 {
		return ErrBookNotFound
	}

	return nil
}
