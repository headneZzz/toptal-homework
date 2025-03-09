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

type DB struct {
	*sqlx.DB
}

func Connect(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	return &DB{db}, nil
}

func NewDB(db *sqlx.DB) *DB {
	return &DB{db}
}

func (r *DB) QueryRow(ctx context.Context, operation string, query string, args ...interface{}) *sqlx.Row {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.QueryRowxContext(ctx, query, args...)
}

func (r *DB) Query(ctx context.Context, operation string, query string, args ...interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.QueryxContext(ctx, query, args...)
}

func (r *DB) Exec(ctx context.Context, operation string, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.ExecContext(ctx, query, args...)
}

func (r *DB) NamedExec(ctx context.Context, operation string, query string, arg interface{}) (sql.Result, error) {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.NamedExecContext(ctx, query, arg)
}

func (r *DB) Get(ctx context.Context, operation string, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.GetContext(ctx, dest, query, args...)
}

func (r *DB) Select(ctx context.Context, operation string, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() {
		metrics.DatabaseQueryDuration.
			WithLabelValues(operation).
			Observe(time.Since(start).Seconds())
	}()

	return r.DB.SelectContext(ctx, dest, query, args...)
}

func (r *DB) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				err = fmt.Errorf("failed to commit transaction: %w", err)
			}
		}
	}()

	err = fn(tx)
	return err
}
