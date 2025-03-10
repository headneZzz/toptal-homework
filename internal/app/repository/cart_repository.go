package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"toptal/internal/app/config"
	"toptal/internal/app/domain"
	"toptal/internal/app/repository/model"
	"toptal/internal/pkg/pg"

	"github.com/jmoiron/sqlx"
)

const (
	sqlGetCart = `
  		SELECT b.id, b.title, b.author, b.year, b.price, b.stock, b.category_id
  		FROM books b
  		JOIN cart_items ci ON b.id = ci.book_id
  		JOIN cart c ON ci.cart_id = c.id
  		WHERE c.user_id = $1
	`
	sqlSelectBookStock    = `SELECT stock FROM books WHERE id = $1 FOR UPDATE`
	sqlCheckItemInCart    = `SELECT COUNT(1) FROM cart_items WHERE cart_id = $1 AND book_id = $2`
	sqlInsertCartItem     = `INSERT INTO cart_items (cart_id, book_id, updated_at) VALUES ($1, $2, now())`
	sqlUpdateCartItemTime = `UPDATE cart_items SET updated_at = now() WHERE cart_id = $1 AND book_id = $2`
	sqlRemoveFromCart     = `DELETE FROM cart_items WHERE cart_id = $1 AND book_id = $2`
	sqlGetBooksInCart     = `SELECT book_id FROM cart_items WHERE cart_id = $1 FOR UPDATE`
	sqlUpdateBookStock    = `UPDATE books SET stock = stock - 1 WHERE id = $1 AND stock > 0 RETURNING stock`
	sqlGetCartByUser      = `SELECT id FROM cart WHERE user_id = $1`
	sqlInsertCart         = `INSERT INTO cart (user_id, updated_at) VALUES ($1, now()) RETURNING id`
	sqlUpdateCartTime     = `UPDATE cart SET updated_at = now() WHERE id = $1`
	sqlClearCartItems     = `DELETE FROM cart_items WHERE cart_id = $1`
	sqlFindExpiredCarts   = `SELECT id FROM cart WHERE updated_at < $1`
	sqlDeleteCart         = `DELETE FROM cart WHERE id = $1`
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

// ensureCart checks if a cart exists for the given userId.
// If no cart exists, it creates a new cart and returns its id.
func (r *CartRepository) ensureCart(ctx context.Context, tx *sqlx.Tx, userId int) (int, error) {
	var cartId int
	err := tx.GetContext(ctx, &cartId, sqlGetCartByUser, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.GetContext(ctx, &cartId, sqlInsertCart, userId)
			if err != nil {
				return 0, model.WrapDatabaseError(err, "failed to create cart")
			}
		} else {
			return 0, model.WrapDatabaseError(err, "failed to get cart")
		}
	}
	// Update the cart timestamp to reflect recent activity.
	_, err = tx.ExecContext(ctx, sqlUpdateCartTime, cartId)
	if err != nil {
		return 0, model.WrapDatabaseError(err, "failed to update cart timestamp")
	}
	return cartId, nil
}

func (r *CartRepository) GetCart(ctx context.Context, userId int) ([]domain.Book, error) {
	var books []model.Book
	err := r.db.Select(ctx, "get_cart", &books, sqlGetCart, userId)
	if err != nil {
		return nil, model.WrapDatabaseError(err, "failed to get cart")
	}
	return toDomainBooks(books), nil
}

func (r *CartRepository) AddToCart(ctx context.Context, userId int, bookId int) error {
	return r.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		cartId, err := r.ensureCart(ctx, tx, userId)
		if err != nil {
			return fmt.Errorf("failed to ensure cart: %w", err)
		}
		if err := r.checkBookAvailability(ctx, tx, bookId); err != nil {
			return fmt.Errorf("book not available: %w", err)
		}
		if err := r.addOrUpdateCartItem(ctx, tx, cartId, bookId); err != nil {
			return fmt.Errorf("failed to add book to cart: %w", err)
		}
		return nil
	})
}

func (r *CartRepository) checkBookAvailability(ctx context.Context, tx *sqlx.Tx, bookId int) error {
	var stock int
	err := tx.GetContext(ctx, &stock, sqlSelectBookStock, bookId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrBookNotFound
		}
		return model.WrapDatabaseError(err, "failed to get book stock")
	}

	if stock <= 0 {
		return domain.ErrBookOutOfStock
	}

	return nil
}

func (r *CartRepository) addOrUpdateCartItem(ctx context.Context, tx *sqlx.Tx, cartId int, bookId int) error {
	var count int
	err := tx.GetContext(ctx, &count, sqlCheckItemInCart, cartId, bookId)
	if err != nil {
		return model.WrapDatabaseError(err, "failed to check if book already in cart")
	}

	if count > 0 {
		_, err = tx.ExecContext(ctx, sqlUpdateCartItemTime, cartId, bookId)
		if err != nil {
			return model.WrapDatabaseError(err, "failed to update cart item timestamp")
		}
		return nil
	}

	_, err = tx.ExecContext(ctx, sqlInsertCartItem, cartId, bookId)
	if err != nil {
		return model.WrapDatabaseError(err, "failed to add book to cart")
	}

	return nil
}

func (r *CartRepository) RemoveFromCart(ctx context.Context, userId int, bookId int) error {
	cartId, err := r.getCartId(ctx, userId)
	if err != nil {
		return fmt.Errorf("failed to get user cart: %w", err)
	}

	result, err := r.db.Exec(ctx, "remove_from_cart", sqlRemoveFromCart, cartId, bookId)
	if err != nil {
		return model.WrapDatabaseError(err, "failed to remove book from cart")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return model.WrapDatabaseError(err, "failed to get affected rows")
	}

	if rows == 0 {
		return domain.ErrBookNotInCart
	}

	slog.Info("Book removed from cart", "user_id", userId, "book_id", bookId)
	return nil
}

// getCartId fetches the cart id for the given user.
func (r *CartRepository) getCartId(ctx context.Context, userId int) (int, error) {
	var cartId int
	err := r.db.Get(ctx, "get_cart_by_user", &cartId, sqlGetCartByUser, userId)
	if err != nil {
		return 0, err
	}
	return cartId, nil
}

func (r *CartRepository) Purchase(ctx context.Context, userId int) error {
	return r.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		cartId, err := r.ensureCart(ctx, tx, userId)
		if err != nil {
			return fmt.Errorf("failed to ensure cart: %w", err)
		}

		rows, err := tx.QueryxContext(ctx, sqlGetBooksInCart, cartId)
		if err != nil {
			return model.WrapDatabaseError(err, "failed to get books in cart")
		}
		defer func(rows *sqlx.Rows) {
			err := rows.Close()
			if err != nil {
				slog.Error("failed to close rows", "error", err)
			}
		}(rows)

		var bookIds []int
		for rows.Next() {
			var bookId int
			if err := rows.Scan(&bookId); err != nil {
				return model.WrapDatabaseError(err, "failed to scan book id")
			}
			bookIds = append(bookIds, bookId)
		}

		if len(bookIds) == 0 {
			return domain.ErrCartEmpty
		}

		for _, bookId := range bookIds {
			var remainingStock int
			err := tx.GetContext(ctx, &remainingStock, sqlUpdateBookStock, bookId)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) || remainingStock < 0 {
					return domain.ErrBookOutOfStock
				}
				return model.WrapDatabaseError(err, fmt.Sprintf("failed to update book stock for book %d", bookId))
			}
		}

		_, err = tx.ExecContext(ctx, sqlClearCartItems, cartId)
		if err != nil {
			return model.WrapDatabaseError(err, "failed to clear cart items")
		}
		_, err = tx.ExecContext(ctx, sqlDeleteCart, cartId)
		if err != nil {
			return model.WrapDatabaseError(err, fmt.Sprintf("failed to delete cart %d", cartId))
		}

		slog.Info("Purchase completed", "user_id", userId, "books_count", len(bookIds))
		return nil
	})
}

func (r *CartRepository) CleanExpiredCarts(ctx context.Context) error {
	return r.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		expirationTime := time.Now().Add(-r.cartConfig.ExpiryTime)

		rows, err := tx.QueryxContext(ctx, sqlFindExpiredCarts, expirationTime)
		if err != nil {
			return model.WrapDatabaseError(err, "failed to find expired carts")
		}
		defer func(rows *sqlx.Rows) {
			err := rows.Close()
			if err != nil {
				slog.Error("failed to close rows", "error", err)
			}
		}(rows)

		var cartIds []int
		for rows.Next() {
			var cartId int
			if err := rows.Scan(&cartId); err != nil {
				return model.WrapDatabaseError(err, "failed to scan cart id")
			}
			cartIds = append(cartIds, cartId)
		}

		if len(cartIds) == 0 {
			slog.Debug("No expired carts found")
			return nil
		}

		for _, cartId := range cartIds {
			_, err := tx.ExecContext(ctx, sqlClearCartItems, cartId)
			if err != nil {
				return model.WrapDatabaseError(err, fmt.Sprintf("failed to clear cart items for cart %d", cartId))
			}
			_, err = tx.ExecContext(ctx, sqlDeleteCart, cartId)
			if err != nil {
				return model.WrapDatabaseError(err, fmt.Sprintf("failed to delete expired cart %d", cartId))
			}
			slog.Info("Cleared expired cart", "cart_id", cartId)
		}

		slog.Info("Cleaned expired carts", "carts_count", len(cartIds))
		return nil
	})
}
