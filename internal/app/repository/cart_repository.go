package repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"toptal/internal/app/config"
	"toptal/internal/app/domain"
	"toptal/internal/pkg/pg"
)

const (
	sqlGetCart = `
		SELECT b.id, b.title, b.author, b.year, b.price, b.stock, b.category_id
		FROM books b
		JOIN cart c ON b.id = c.book_id
		WHERE c.user_id = $1
		`
	sqlSelectBookStock      = `SELECT stock FROM books WHERE id = $1 FOR UPDATE`
	sqlCheckItemInCart      = `SELECT COUNT(1) FROM cart WHERE user_id = $1 AND book_id = $2`
	sqlInsertCartItem       = `INSERT INTO cart (user_id, book_id, updated_at) VALUES ($1, $2, now())`
	sqlUpdateCartItemTime   = `UPDATE cart SET updated_at = now() WHERE user_id = $1 AND book_id = $2`
	sqlRemoveFromCart       = `DELETE FROM cart WHERE user_id = $1 AND book_id = $2`
	sqlGetBooksInCart       = `SELECT book_id FROM cart WHERE user_id = $1 FOR UPDATE`
	sqlUpdateBookStock      = `UPDATE books SET stock = stock - 1 WHERE id = $1 AND stock > 0`
	sqlUpdateUserUpdateTime = `UPDATE users SET updated_at = now() WHERE id = $1`
	sqlClearCart            = `DELETE FROM cart WHERE user_id = $1`
	sqlFindExpiredUsers     = `
		SELECT id FROM users 
		WHERE updated_at < $1 
		AND EXISTS (SELECT 1 FROM cart WHERE cart.user_id = users.id)
		`
)

type CartRepository struct {
	db         *pg.DB
	cartConfig *config.CartConfig
}

func NewCartRepository(db *pg.DB, cartConfig *config.CartConfig) *CartRepository {
	return &CartRepository{
		db:         db,
		cartConfig: cartConfig,
	}
}

func (r *CartRepository) GetCart(ctx context.Context, userId int) ([]domain.Book, error) {
	var books []domain.Book
	err := r.db.Select(ctx, "get_cart", &books, sqlGetCart, userId)
	if err != nil {
		return nil, WrapDatabaseError(err, "failed to get cart")
	}
	return books, nil
}

func (r *CartRepository) AddToCart(ctx context.Context, userId int, bookId int) error {
	return r.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		if err := r.checkBookAvailability(ctx, tx, bookId); err != nil {
			return err
		}
		if err := r.addOrUpdateCartItem(ctx, tx, userId, bookId); err != nil {
			return err
		}
		return nil
	})
}

func (r *CartRepository) checkBookAvailability(ctx context.Context, tx *sqlx.Tx, bookId int) error {
	var stock int
	err := tx.GetContext(ctx, &stock, sqlSelectBookStock, bookId)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return fmt.Errorf("book does not exist")
		}
		return WrapDatabaseError(err, "failed to get book stock")
	}

	if stock <= 0 {
		return ErrBookOutOfStock
	}

	return nil
}

func (r *CartRepository) addOrUpdateCartItem(ctx context.Context, tx *sqlx.Tx, userId int, bookId int) error {
	var count int
	err := tx.GetContext(ctx, &count, sqlCheckItemInCart, userId, bookId)
	if err != nil {
		return WrapDatabaseError(err, "failed to check if book already in cart")
	}

	if count > 0 {
		_, err = tx.ExecContext(ctx, sqlUpdateCartItemTime, userId, bookId)
		if err != nil {
			return WrapDatabaseError(err, "failed to update cart item timestamp")
		}
		return nil
	}

	_, err = tx.ExecContext(ctx, sqlInsertCartItem, userId, bookId)
	if err != nil {
		return WrapDatabaseError(err, "failed to add book to cart")
	}

	_, err = tx.ExecContext(ctx, sqlUpdateUserUpdateTime, userId)
	if err != nil {
		return WrapDatabaseError(err, "failed to update user cart timestamp")
	}

	return nil
}

func (r *CartRepository) RemoveFromCart(ctx context.Context, userId int, bookId int) error {
	result, err := r.db.Exec(ctx, "remove_from_cart", sqlRemoveFromCart, userId, bookId)
	if err != nil {
		return WrapDatabaseError(err, "failed to remove book from cart")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return WrapDatabaseError(err, "failed to get affected rows")
	}

	if rows == 0 {
		return ErrBookNotInCart
	}

	slog.Info("Book removed from cart", "user_id", userId, "book_id", bookId)
	return nil
}

func (r *CartRepository) Purchase(ctx context.Context, userId int) error {
	return r.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		rows, err := tx.QueryxContext(ctx, sqlGetBooksInCart, userId)
		if err != nil {
			return WrapDatabaseError(err, "failed to get books in cart")
		}
		defer rows.Close()

		var bookIds []int
		for rows.Next() {
			var bookId int
			if err := rows.Scan(&bookId); err != nil {
				return WrapDatabaseError(err, "failed to scan book id")
			}
			bookIds = append(bookIds, bookId)
		}

		if len(bookIds) == 0 {
			return ErrCartEmpty
		}

		for _, bookId := range bookIds {
			result, err := tx.ExecContext(ctx, sqlUpdateBookStock, bookId)
			if err != nil {
				return WrapDatabaseError(err, fmt.Sprintf("failed to update book stock for book %d", bookId))
			}

			affected, err := result.RowsAffected()
			if err != nil {
				return WrapDatabaseError(err, "failed to get affected rows")
			}

			if affected == 0 {
				return ErrBookOutOfStock
			}
		}

		_, err = tx.ExecContext(ctx, sqlClearCart, userId)
		if err != nil {
			return WrapDatabaseError(err, "failed to clear cart")
		}

		slog.Info("Purchase completed", "user_id", userId, "books_count", len(bookIds))
		return nil
	})
}

func (r *CartRepository) CleanExpiredCarts(ctx context.Context) error {
	return r.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		expirationTime := time.Now().Add(-r.cartConfig.ExpiryTime)

		rows, err := tx.QueryxContext(ctx, sqlFindExpiredUsers, expirationTime)
		if err != nil {
			return WrapDatabaseError(err, "failed to find users with expired carts")
		}
		defer rows.Close()

		var userIds []int
		for rows.Next() {
			var userId int
			if err := rows.Scan(&userId); err != nil {
				return WrapDatabaseError(err, "failed to scan user id")
			}
			userIds = append(userIds, userId)
		}

		if len(userIds) == 0 {
			slog.Debug("No expired carts found")
			return nil
		}

		var totalRemoved int64
		for _, userId := range userIds {
			var itemsCount int
			err := tx.GetContext(ctx, &itemsCount, `SELECT COUNT(*) FROM cart WHERE user_id = $1`, userId)
			if err != nil {
				return WrapDatabaseError(err, "failed to count cart items")
			}

			result, err := tx.ExecContext(ctx, sqlClearCart, userId)
			if err != nil {
				return WrapDatabaseError(err, fmt.Sprintf("failed to clear cart for user %d", userId))
			}

			removed, err := result.RowsAffected()
			if err != nil {
				return WrapDatabaseError(err, "failed to get affected rows")
			}

			totalRemoved += removed
			slog.Info("Cleared expired cart", "user_id", userId, "items_removed", removed)
		}

		slog.Info("Cleaned expired carts", "users_count", len(userIds), "total_items_removed", totalRemoved)
		return nil
	})
}
