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
	sqlSelectBookStock        = `SELECT stock FROM books WHERE id = $1 FOR UPDATE`
	sqlCheckItemInCart        = `SELECT COUNT(1) FROM cart_items WHERE cart_id = $1 AND book_id = $2`
	sqlUpdateCartItemTime     = `UPDATE cart_items SET updated_at = now() WHERE cart_id = $1 AND book_id = $2`
	sqlInsertCartItem         = `INSERT INTO cart_items (cart_id, book_id, updated_at) VALUES ($1, $2, now())`
	sqlRemoveFromCart         = `DELETE FROM cart_items WHERE cart_id = $1 AND book_id = $2`
	sqlGetCartByUser          = `SELECT id FROM cart WHERE user_id = $1`
	sqlInsertCart             = `INSERT INTO cart (user_id, updated_at) VALUES ($1, now()) RETURNING id`
	sqlUpdateCartTime         = `UPDATE cart SET updated_at = now() WHERE id = $1`
	sqlClearCartItems         = `DELETE FROM cart_items WHERE cart_id = $1`
	sqlDeleteCart             = `DELETE FROM cart WHERE id = $1`
	sqlDeleteExpiredCartItems = `DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM cart WHERE updated_at < $1)`
	sqlDeleteExpiredCarts     = `DELETE FROM cart WHERE updated_at < $1`
	sqlSelectCartItemsCount   = `SELECT COUNT(*) FROM cart_items WHERE cart_id = $1`
	sqlUpdateBooksStock       = `
		UPDATE books
		SET stock = stock - 1
		WHERE id IN (
			SELECT book_id FROM cart_items WHERE cart_id = $1
		) AND stock > 0
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

// ensureCart checks if a cart exists for the given userId.
// If no cart exists, it creates a new cart and returns its id.
func (r *CartRepository) ensureCart(ctx context.Context, tx *sqlx.Tx, userId int) (int, error) {
	cartId, err := r.getCartId(ctx, userId)
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

		var totalItems int64
		if err := tx.GetContext(ctx, &totalItems, sqlSelectCartItemsCount, cartId); err != nil {
			return model.WrapDatabaseError(err, "failed to count cart items")
		}
		if totalItems == 0 {
			return domain.ErrCartEmpty
		}

		result, err := tx.ExecContext(ctx, sqlUpdateBooksStock, cartId)
		if err != nil {
			return model.WrapDatabaseError(err, "failed to update book stock")
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return model.WrapDatabaseError(err, "failed to get affected rows")
		}
		if rows != totalItems {
			return domain.ErrBookOutOfStock
		}

		// clear cart
		if _, err := tx.ExecContext(ctx, sqlClearCartItems, cartId); err != nil {
			return model.WrapDatabaseError(err, "failed to clear cart items")
		}
		if _, err := tx.ExecContext(ctx, sqlDeleteCart, cartId); err != nil {
			return model.WrapDatabaseError(err, fmt.Sprintf("failed to delete cart %d", cartId))
		}

		slog.Info("Purchase completed", "user_id", userId, "books_count", totalItems)
		return nil
	})
}

func (r *CartRepository) CleanExpiredCarts(ctx context.Context) error {
	return r.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		expirationTime := time.Now().Add(-r.cartConfig.ExpiryTime)
		_, err := tx.ExecContext(ctx, sqlDeleteExpiredCartItems, expirationTime)
		if err != nil {
			return model.WrapDatabaseError(err, "failed to clear cart items for expired carts")
		}

		res, err := tx.ExecContext(ctx, sqlDeleteExpiredCarts, expirationTime)
		if err != nil {
			return model.WrapDatabaseError(err, "failed to delete expired carts")
		}
		affected, _ := res.RowsAffected()
		slog.Info(fmt.Sprintf("Cleaned expired carts, carts deleted: %d", affected))
		return nil
	})
}
