package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"toptal/internal/app/config"
	"toptal/internal/pkg/pg"
)

func setupCartTest(t *testing.T) (*CartRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	pgDB := pg.NewDB(sqlxDB)

	cartConfig := &config.CartConfig{
		CleanupInterval: time.Minute,
		ExpiryTime:      30 * time.Minute,
	}

	repo := NewCartRepository(pgDB, cartConfig)
	return repo, mock
}

func TestCartRepository_GetCart(t *testing.T) {
	repo, mock := setupCartTest(t)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "author", "year", "price", "stock", "category_id"}).
			AddRow(1, "Book 1", "Author 1", 2020, 1000, 5, 1).
			AddRow(2, "Book 2", "Author 2", 2021, 2000, 3, 2)

		mock.ExpectQuery(`SELECT b\.id, b\.title, b\.author, b\.year, b\.price, b\.stock, b\.category_id`).
			WithArgs(1).
			WillReturnRows(rows)

		books, err := repo.GetCart(context.Background(), 1)
		assert.NoError(t, err)
		assert.Len(t, books, 2)
		assert.Equal(t, "Book 1", books[0].Title())
		assert.Equal(t, "Book 2", books[1].Title())
	})

	t.Run("Empty cart", func(t *testing.T) {
		mock.ExpectQuery(`SELECT b\.id, b\.title, b\.author, b\.year, b\.price, b\.stock, b\.category_id`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "year", "price", "stock", "category_id"}))

		books, err := repo.GetCart(context.Background(), 1)
		assert.NoError(t, err)
		assert.Empty(t, books)
	})
}

func TestCartRepository_AddToCart(t *testing.T) {
	repo, mock := setupCartTest(t)

	t.Run("Success - new item", func(t *testing.T) {
		mock.ExpectBegin()
		// Ensure cart exists
		mock.ExpectQuery(`SELECT id FROM cart WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		// Expect update cart timestamp Exec
		mock.ExpectExec(`UPDATE cart SET updated_at = now\(\) WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		// Check book stock
		mock.ExpectQuery(`SELECT stock FROM books WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(5))
		// Check if book already in cart_items
		mock.ExpectQuery(`SELECT COUNT\(1\) FROM cart_items WHERE cart_id = \$1 AND book_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		// Insert into cart_items for new item
		mock.ExpectExec(`INSERT INTO cart_items`).
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.AddToCart(context.Background(), 1, 1)
		assert.NoError(t, err)
	})

	t.Run("Book out of stock", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT id FROM cart WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`UPDATE cart SET updated_at = now\(\) WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT stock FROM books WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(0))
		mock.ExpectRollback()

		err := repo.AddToCart(context.Background(), 1, 1)
		assert.Error(t, err)
	})

	t.Run("Book already in cart - update timestamp", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT id FROM cart WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`UPDATE cart SET updated_at = now\(\) WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT stock FROM books WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(5))
		mock.ExpectQuery(`SELECT COUNT\(1\) FROM cart_items WHERE cart_id = \$1 AND book_id = \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		// Update the existing cart item timestamp
		mock.ExpectExec(`UPDATE cart_items SET updated_at = now\(\) WHERE cart_id = \$1 AND book_id = \$2`).
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.AddToCart(context.Background(), 1, 1)
		assert.NoError(t, err)
	})
}

func TestCartRepository_RemoveFromCart(t *testing.T) {
	repo, mock := setupCartTest(t)

	t.Run("Success", func(t *testing.T) {
		// Ensure cart exists
		mock.ExpectQuery(`SELECT id FROM cart WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = \$1 AND book_id = \$2`).
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.RemoveFromCart(context.Background(), 1, 1)
		assert.NoError(t, err)
	})

	t.Run("Book not in cart", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id FROM cart WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = \$1 AND book_id = \$2`).
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.RemoveFromCart(context.Background(), 1, 1)
		assert.Error(t, err)
	})
}

func TestCartRepository_Purchase(t *testing.T) {
	repo, mock := setupCartTest(t)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT id FROM cart WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`UPDATE cart SET updated_at = now\(\) WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		mock.ExpectExec(`UPDATE books\s+SET stock = stock - 1\s+WHERE id IN \(\s*SELECT book_id FROM cart_items WHERE cart_id = \$1\s*\) AND stock > 0`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 2))
		mock.ExpectExec(`DELETE FROM cart WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Purchase(context.Background(), 1)
		assert.NoError(t, err)
	})

	t.Run("Empty cart", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT id FROM cart WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`UPDATE cart SET updated_at = now\(\) WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}))
		mock.ExpectRollback()

		err := repo.Purchase(context.Background(), 1)
		assert.Error(t, err)
	})

	t.Run("Book out of stock during purchase", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT id FROM cart WHERE user_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`UPDATE cart SET updated_at = now\(\) WHERE id = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM cart_items WHERE cart_id = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectExec(`UPDATE books\s+SET stock = stock - 1\s+WHERE id IN \(\s*SELECT book_id FROM cart_items WHERE cart_id = \$1\s*\) AND stock > 0`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectRollback()

		err := repo.Purchase(context.Background(), 1)
		assert.Error(t, err)
	})
}

func TestCartRepository_CleanExpiredCarts(t *testing.T) {
	t.Run("Success - clean expired carts", func(t *testing.T) {
		repo, mock := setupCartTest(t)
		mock.ExpectBegin()
		// Expect deletion of cart items for expired carts.
		mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id IN \(SELECT id FROM cart WHERE updated_at < \$1\)`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 2))
		// Expect deletion of expired carts.
		mock.ExpectExec(`DELETE FROM cart WHERE updated_at < \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		err := repo.CleanExpiredCarts(context.Background())
		assert.NoError(t, err)
	})
}
