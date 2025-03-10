package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
	"toptal/internal/app/repository/model"

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

		mock.ExpectQuery("SELECT b.id, b.title, b.author, b.year, b.price, b.stock, b.category_id").
			WithArgs(1).
			WillReturnRows(rows)

		books, err := repo.GetCart(context.Background(), 1)
		assert.NoError(t, err)
		assert.Len(t, books, 2)
		assert.Equal(t, "Book 1", books[0].Title)
		assert.Equal(t, "Book 2", books[1].Title)
	})

	t.Run("Empty cart", func(t *testing.T) {
		mock.ExpectQuery("SELECT b.id, b.title, b.author, b.year, b.price, b.stock, b.category_id").
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

		// Check book stock
		mock.ExpectQuery("SELECT stock FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(5))

		// Check if book in cart
		mock.ExpectQuery("SELECT COUNT\\(1\\) FROM cart WHERE user_id = \\$1 AND book_id = \\$2").
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Insert into cart
		mock.ExpectExec("INSERT INTO cart").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Update user timestamp
		mock.ExpectExec("UPDATE users SET updated_at").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.AddToCart(context.Background(), 1, 1)
		assert.NoError(t, err)
	})

	t.Run("Book out of stock", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery("SELECT stock FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(0))

		mock.ExpectRollback()

		err := repo.AddToCart(context.Background(), 1, 1)
		assert.Error(t, err)
	})

	t.Run("Book already in cart - update timestamp", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery("SELECT stock FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(5))

		mock.ExpectQuery("SELECT COUNT\\(1\\) FROM cart WHERE user_id = \\$1 AND book_id = \\$2").
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectExec("UPDATE cart SET updated_at").
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
		mock.ExpectExec("DELETE FROM cart WHERE user_id = \\$1 AND book_id = \\$2").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.RemoveFromCart(context.Background(), 1, 1)
		assert.NoError(t, err)
	})

	t.Run("Book not in cart", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM cart WHERE user_id = \\$1 AND book_id = \\$2").
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

		// Get books in cart
		mock.ExpectQuery("SELECT book_id FROM cart WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"book_id"}).
				AddRow(1).
				AddRow(2))

		// Update stock for each book
		mock.ExpectExec("UPDATE books SET stock = stock - 1 WHERE id = \\$1").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE books SET stock = stock - 1 WHERE id = \\$1").
			WithArgs(2).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Clear cart
		mock.ExpectExec("DELETE FROM cart WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 2))

		mock.ExpectCommit()

		err := repo.Purchase(context.Background(), 1)
		assert.NoError(t, err)
	})

	t.Run("Empty cart", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery("SELECT book_id FROM cart WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"book_id"}))

		mock.ExpectRollback()

		err := repo.Purchase(context.Background(), 1)
		assert.Error(t, err)
	})

	t.Run("Book out of stock during purchase", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery("SELECT book_id FROM cart WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"book_id"}).AddRow(1))

		mock.ExpectExec("UPDATE books SET stock = stock - 1 WHERE id = \\$1").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectRollback()

		err := repo.Purchase(context.Background(), 1)
		assert.Error(t, err)
	})
}

func TestCartRepository_CleanExpiredCarts(t *testing.T) {
	repo, mock := setupCartTest(t)

	t.Run("Success - clean expired carts", func(t *testing.T) {
		mock.ExpectBegin()

		// Find expired users
		mock.ExpectQuery("SELECT id FROM users WHERE updated_at < \\$1").
			WithArgs(sqlmock.AnyArg()). // We use AnyArg() because the time will be dynamic
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1).
				AddRow(2))

		// Count items for each user
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM cart WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Clear cart for first user
		mock.ExpectExec("DELETE FROM cart WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 2))

		// Count items for second user
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM cart WHERE user_id = \\$1").
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		// Clear cart for second user
		mock.ExpectExec("DELETE FROM cart WHERE user_id = \\$1").
			WithArgs(2).
			WillReturnResult(sqlmock.NewResult(1, 3))

		mock.ExpectCommit()

		err := repo.CleanExpiredCarts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("No expired carts", func(t *testing.T) {
		mock.ExpectBegin()

		// Find expired users with the correct WHERE clause
		mock.ExpectQuery("SELECT id FROM users WHERE updated_at < \\$1").
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		mock.ExpectCommit()

		err := repo.CleanExpiredCarts(context.Background())
		assert.NoError(t, err)
	})
}

func TestCartRepository_Purchase_ConcurrencyCheck(t *testing.T) {
	repo, mock := setupCartTest(t)

	t.Run("No lost updates during concurrent purchases", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errs := make([]error, 3)

		mock.MatchExpectationsInOrder(false)

		// First successful purchase
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT book_id FROM cart WHERE user_id = \\$1 FOR UPDATE").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"book_id"}).AddRow(1))
		mock.ExpectQuery("UPDATE books SET stock = stock - 1 WHERE id = \\$1 AND stock > 0 RETURNING stock").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}).AddRow(2))
		mock.ExpectExec("DELETE FROM cart WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Second purchase (out of stock)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT book_id FROM cart WHERE user_id = \\$1 FOR UPDATE").
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"book_id"}).AddRow(1))
		mock.ExpectQuery("UPDATE books SET stock = stock - 1 WHERE id = \\$1 AND stock > 0 RETURNING stock").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}))
		mock.ExpectRollback()

		// Third purchase (out of stock)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT book_id FROM cart WHERE user_id = \\$1 FOR UPDATE").
			WithArgs(3).
			WillReturnRows(sqlmock.NewRows([]string{"book_id"}).AddRow(1))
		mock.ExpectQuery("UPDATE books SET stock = stock - 1 WHERE id = \\$1 AND stock > 0 RETURNING stock").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"stock"}))
		mock.ExpectRollback()

		wg.Add(3)
		for i := 1; i <= 3; i++ {
			userID := i
			go func(id int) {
				defer wg.Done()
				err := repo.Purchase(context.Background(), id)
				mu.Lock()
				errs[id-1] = err
				mu.Unlock()
			}(userID)
		}

		wg.Wait()

		successCount := 0
		failCount := 0
		for i, err := range errs {
			if err == nil {
				successCount++
				t.Logf("Purchase %d succeeded", i+1)
			} else if errors.Is(err, model.ErrBookOutOfStock) {
				failCount++
				t.Logf("Purchase %d failed with out of stock error", i+1)
			} else {
				t.Errorf("Purchase %d failed with unexpected error: %v", i+1, err)
			}
		}

		assert.Equal(t, 1, successCount, "expected 1 successful purchase")
		assert.Equal(t, 2, failCount, "expected 2 failed purchases")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
