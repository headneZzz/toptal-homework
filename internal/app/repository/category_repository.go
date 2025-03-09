package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log/slog"
	"toptal/internal/app/domain"
	"toptal/internal/pkg/pg"
)

const (
	sqlFindCategoryById = `SELECT * FROM categories WHERE id = $1`
	sqlFindCategories   = `SELECT * FROM categories`
	sqlInsertCategory   = `INSERT INTO categories (name) VALUES ($1)`
	sqlUpdateCategory   = `UPDATE categories SET name = $1 WHERE id = $2`
	sqlDeleteCategory   = `DELETE FROM categories WHERE id = $1`
)

const UniqueViolationErr = "23505"

type CategoryRepository struct {
	db *pg.DB
}

func NewCategoryRepository(db *pg.DB) *CategoryRepository {
	return &CategoryRepository{db}
}

func (r *CategoryRepository) FindCategoryById(ctx context.Context, id int) (domain.Category, error) {
	var category domain.Category
	row := r.db.QueryRow(ctx, "find_category_by_id", sqlFindCategoryById, id)
	err := row.StructScan(&category)
	return category, err
}

func (r *CategoryRepository) FindCategories(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.Select(ctx, "find_categories", &categories, sqlFindCategories)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) InsertCategory(ctx context.Context, book domain.Category) error {
	result, err := r.db.Exec(ctx, "insert_category", sqlInsertCategory, book.Name)
	if err != nil {
		var pqErr *pq.Error
		if ok := errors.As(err, &pqErr); ok && pqErr.Code == UniqueViolationErr {
			return fmt.Errorf("category already exists")
		}
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		return err
	}
	slog.Info("CategoryRepository.InsertCategory", "affect", affect)
	return nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, category domain.Category) error {
	result, err := r.db.Exec(ctx, "update_category", sqlUpdateCategory, category.Name, category.Id)
	if err != nil {
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		return err
	}
	slog.Info("CategoryRepository.UpdateCategory", "affect", affect)
	return nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx, "delete_category", sqlDeleteCategory, id)
	if err != nil {
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		return err
	}
	slog.Info("CategoryRepository.DeleteCategory", "affect", affect)
	return nil
}
