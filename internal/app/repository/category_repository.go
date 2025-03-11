package repository

import (
	"context"
	"fmt"
	"log/slog"
	"toptal/internal/app/domain"
	"toptal/internal/app/repository/model"
	"toptal/internal/pkg/pg"
)

const (
	sqlFindCategoryById = `SELECT * FROM categories WHERE id = $1`
	sqlFindCategories   = `SELECT * FROM categories`
	sqlInsertCategory   = `INSERT INTO categories (name) VALUES ($1)`
	sqlUpdateCategory   = `UPDATE categories SET name = $1 WHERE id = $2`
	sqlDeleteCategory   = `DELETE FROM categories WHERE id = $1`
)

type CategoryRepository struct {
	db *pg.DB
}

func NewCategoryRepository(db *pg.DB) *CategoryRepository {
	return &CategoryRepository{db}
}

func (r *CategoryRepository) FindCategoryById(ctx context.Context, id int) (domain.Category, error) {
	var category model.Category
	row := r.db.QueryRow(ctx, "find_category_by_id", sqlFindCategoryById, id)
	err := row.StructScan(&category)
	if err != nil {
		return domain.Category{}, fmt.Errorf("failed to find category by id: %w", err)
	}
	return toDomainCategory(category)
}

func (r *CategoryRepository) FindCategories(ctx context.Context) ([]domain.Category, error) {
	var categories []model.Category
	err := r.db.Select(ctx, "find_categories", &categories, sqlFindCategories)
	if err != nil {
		return nil, fmt.Errorf("failed to find categories: %w", err)
	}
	return toDomainCategories(categories)
}

func (r *CategoryRepository) InsertCategory(ctx context.Context, category domain.Category) error {
	result, err := r.db.Exec(ctx, "insert_category", sqlInsertCategory, category.Name())
	if err != nil {
		if pg.IsUniqueViolationErr(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to insert category: %w", err)
	}
	affect, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	slog.Info("CategoryRepository.InsertCategory", "affect", affect)
	return nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, category domain.Category) error {
	result, err := r.db.Exec(ctx, "update_category", sqlUpdateCategory, category.Name(), category.Id())
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	affect, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	slog.Info("CategoryRepository.UpdateCategory", "affect", affect)
	return nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx, "delete_category", sqlDeleteCategory, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	affect, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	slog.Info("CategoryRepository.DeleteCategory", "affect", affect)
	return nil
}
