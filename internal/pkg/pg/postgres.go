package pg

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"toptal/internal/app/metrics"

	"github.com/jmoiron/sqlx"
	"toptal/internal/app/config"
)

// DB содержит общую функциональность для всех репозиториев
type DB struct {
	DB *sqlx.DB
}

// Connect NewBaseRepository создает новый базовый репозиторий
func Connect(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	return &DB{
		DB: db,
	}, nil
}

func NewDB(db *sqlx.DB) *DB {
	return &DB{
		DB: db,
	}
}

// QueryRow выполняет запрос и возвращает одну строку с измерением времени выполнения
func (r *DB) QueryRow(ctx context.Context, operation string, query string, args ...interface{}) *sqlx.Row {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.QueryRowxContext(ctx, query, args...)
}

// Query выполняет запрос и возвращает несколько строк с измерением времени выполнения
func (r *DB) Query(ctx context.Context, operation string, query string, args ...interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.QueryxContext(ctx, query, args...)
}

// Exec выполняет запрос, не возвращающий данных, с измерением времени выполнения
func (r *DB) Exec(ctx context.Context, operation string, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.ExecContext(ctx, query, args...)
}

// Get получает одну запись с измерением времени выполнения
func (r *DB) Get(ctx context.Context, operation string, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.GetContext(ctx, dest, query, args...)
}

// Select получает несколько записей с измерением времени выполнения
func (r *DB) Select(ctx context.Context, operation string, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.SelectContext(ctx, dest, query, args...)
}

// BeginTxx начинает транзакцию с измерением времени выполнения
func (r *DB) BeginTxx(ctx context.Context, operation string) (*sqlx.Tx, error) {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation + "_begin_tx").
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.BeginTxx(ctx, nil)
}

// WithTransaction выполняет функцию в транзакции
func (r *DB) WithTransaction(ctx context.Context, operation string, fn func(*sqlx.Tx) error) error {
	tx, err := r.BeginTxx(ctx, operation)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic после rollback
		} else if err != nil {
			_ = tx.Rollback() // Ошибка уже записана в err
		} else {
			err = tx.Commit() // Если ошибок нет, подтверждаем транзакцию
			if err != nil {
				err = fmt.Errorf("failed to commit transaction: %w", err)
			}
		}
	}()

	err = fn(tx)
	return err
}
