package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"

	"toptal/internal/app/config"
	"toptal/internal/app/domain"
	appErrors "toptal/internal/app/errors"
	"toptal/internal/pkg/pg"
)

// SQL запросы для корзины
const (
	sqlGetCart = `
		SELECT b.id, b.title, b.author, b.year, b.price, b.stock, b.category_id
		FROM books b
		JOIN cart c ON b.id = c.book_id
		WHERE c.user_id = $1
		`
	sqlSelectBookStock      = `SELECT stock FROM books WHERE id = $1 FOR UPDATE`
	sqlCheckItemInCart      = `SELECT COUNT(1) FROM cart WHERE user_id = $1 AND book_id = $2`
	sqlInsertCartItem       = `INSERT INTO cart (user_id, book_id, updated_at) VALUES ($1, $2, $3)`
	sqlUpdateCartItemTime   = `UPDATE cart SET updated_at = $3 WHERE user_id = $1 AND book_id = $2`
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

// CartRepository представляет репозиторий для работы с корзиной
type CartRepository struct {
	db         *pg.DB
	cartConfig *config.CartConfig
}

// NewCartRepository создает новый репозиторий для работы с корзиной
func NewCartRepository(db *pg.DB, cartConfig *config.CartConfig) *CartRepository {
	return &CartRepository{
		db:         db,
		cartConfig: cartConfig,
	}
}

// GetCart получает содержимое корзины пользователя
func (r *CartRepository) GetCart(ctx context.Context, userId int) ([]domain.Book, error) {
	var books []domain.Book
	err := r.db.Select(ctx, "get_cart", &books, sqlGetCart, userId)
	if err != nil {
		return nil, appErrors.WrapDatabaseError(err, "failed to get cart")
	}
	return books, nil
}

// AddToCart добавляет книгу в корзину пользователя
func (r *CartRepository) AddToCart(ctx context.Context, userId int, bookId int) error {
	return r.db.WithTransaction(ctx, "add_to_cart", func(tx *sqlx.Tx) error {
		if err := r.checkBookAvailability(ctx, tx, bookId); err != nil {
			return err
		}
		if err := r.addOrUpdateCartItem(ctx, tx, userId, bookId); err != nil {
			return err
		}
		return nil
	})
}

// checkBookAvailability проверяет наличие книги
func (r *CartRepository) checkBookAvailability(ctx context.Context, tx *sqlx.Tx, bookId int) error {
	var stock int
	err := tx.GetContext(ctx, &stock, sqlSelectBookStock, bookId)
	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to get book stock")
	}

	if stock <= 0 {
		return appErrors.ErrBookOutOfStock
	}

	return nil
}

// addOrUpdateCartItem добавляет книгу в корзину или обновляет время добавления
func (r *CartRepository) addOrUpdateCartItem(ctx context.Context, tx *sqlx.Tx, userId int, bookId int) error {
	// Проверяем, есть ли уже эта книга в корзине
	var count int
	err := tx.GetContext(ctx, &count, sqlCheckItemInCart, userId, bookId)
	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to check if book already in cart")
	}

	now := time.Now()

	// Если книга уже в корзине, обновляем время добавления
	if count > 0 {
		_, err = tx.ExecContext(ctx, sqlUpdateCartItemTime, userId, bookId, now)
		if err != nil {
			return appErrors.WrapDatabaseError(err, "failed to update cart item timestamp")
		}
		return nil
	}

	// Добавляем книгу в корзину
	_, err = tx.ExecContext(ctx, sqlInsertCartItem, userId, bookId, now)
	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to add book to cart")
	}

	_, err = tx.ExecContext(ctx, sqlUpdateUserUpdateTime, userId, bookId)
	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to update user cart timestamp")
	}

	return nil
}

// RemoveFromCart удаляет книгу из корзины пользователя
func (r *CartRepository) RemoveFromCart(ctx context.Context, userId int, bookId int) error {
	result, err := r.db.Exec(ctx, "remove_from_cart", sqlRemoveFromCart, userId, bookId)
	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to remove book from cart")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.WrapDatabaseError(err, "failed to get affected rows")
	}

	if rows == 0 {
		return appErrors.ErrBookNotInCart
	}

	slog.Info("Book removed from cart", "user_id", userId, "book_id", bookId)
	return nil
}

// Purchase оформляет покупку всех книг в корзине
func (r *CartRepository) Purchase(ctx context.Context, userId int) error {
	return r.db.WithTransaction(ctx, "purchase", func(tx *sqlx.Tx) error {
		// Получаем список книг в корзине
		rows, err := tx.QueryxContext(ctx, sqlGetBooksInCart, userId)
		if err != nil {
			return appErrors.WrapDatabaseError(err, "failed to get books in cart")
		}
		defer rows.Close()

		var bookIds []int
		for rows.Next() {
			var bookId int
			if err := rows.Scan(&bookId); err != nil {
				return appErrors.WrapDatabaseError(err, "failed to scan book id")
			}
			bookIds = append(bookIds, bookId)
		}

		if len(bookIds) == 0 {
			return appErrors.ErrCartEmpty
		}

		// Обновляем количество книг на складе
		for _, bookId := range bookIds {
			result, err := tx.ExecContext(ctx, sqlUpdateBookStock, bookId)
			if err != nil {
				return appErrors.WrapDatabaseError(err, fmt.Sprintf("failed to update book stock for book %d", bookId))
			}

			affected, err := result.RowsAffected()
			if err != nil {
				return appErrors.WrapDatabaseError(err, "failed to get affected rows")
			}

			if affected == 0 {
				return appErrors.ErrBookOutOfStock
			}
		}

		// Очищаем корзину
		_, err = tx.ExecContext(ctx, sqlClearCart, userId)
		if err != nil {
			return appErrors.WrapDatabaseError(err, "failed to clear cart")
		}

		slog.Info("Purchase completed", "user_id", userId, "books_count", len(bookIds))
		return nil
	})
}

// CleanExpiredCarts удаляет из корзины все книги пользователей, которые не обновляли корзину дольше указанного времени
func (r *CartRepository) CleanExpiredCarts(ctx context.Context) error {
	return r.db.WithTransaction(ctx, "clean_expired_carts", func(tx *sqlx.Tx) error {
		// Находим пользователей с истекшим временем корзины
		expirationTime := time.Now().Add(-r.cartConfig.ExpiryTime)

		// SQL для получения пользователей с истекшим временем корзины
		rows, err := tx.QueryxContext(ctx, sqlFindExpiredUsers, expirationTime)
		if err != nil {
			return appErrors.WrapDatabaseError(err, "failed to find users with expired carts")
		}
		defer rows.Close()

		// Собираем ID пользователей
		var userIds []int
		for rows.Next() {
			var userId int
			if err := rows.Scan(&userId); err != nil {
				return appErrors.WrapDatabaseError(err, "failed to scan user id")
			}
			userIds = append(userIds, userId)
		}

		// Если нет пользователей с истекшим временем, завершаем работу
		if len(userIds) == 0 {
			slog.Debug("No expired carts found")
			return nil
		}

		// Для каждого пользователя с истекшим временем очищаем корзину
		var totalRemoved int64
		for _, userId := range userIds {
			// Подсчитываем количество элементов в корзине
			var itemsCount int
			err := tx.GetContext(ctx, &itemsCount, `SELECT COUNT(*) FROM cart WHERE user_id = $1`, userId)
			if err != nil {
				return appErrors.WrapDatabaseError(err, "failed to count cart items")
			}

			// Удаляем все элементы из корзины
			result, err := tx.ExecContext(ctx, sqlClearCart, userId)
			if err != nil {
				return appErrors.WrapDatabaseError(err, fmt.Sprintf("failed to clear cart for user %d", userId))
			}

			removed, err := result.RowsAffected()
			if err != nil {
				return appErrors.WrapDatabaseError(err, "failed to get affected rows")
			}

			totalRemoved += removed
			slog.Info("Cleared expired cart", "user_id", userId, "items_removed", removed)
		}

		slog.Info("Cleaned expired carts", "users_count", len(userIds), "total_items_removed", totalRemoved)
		return nil
	})
}
