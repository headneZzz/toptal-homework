package service

import (
	"context"
	"toptal/internal/app/domain"
)

type CategoryService struct {
	categoryRepository CategoryRepository
	authService        AuthService
}

func NewCategoryService(categoryRepository CategoryRepository, authService AuthService) *CategoryService {
	return &CategoryService{categoryRepository, authService}
}

func (s *CategoryService) GetCategoryById(ctx context.Context, id int) (domain.Category, error) {
	return s.categoryRepository.FindCategoryById(ctx, id)
}

func (s *CategoryService) GetCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categoryRepository.FindCategories(ctx)
}

func (s *CategoryService) CreateCategory(ctx context.Context, book domain.Category) error {
	return s.categoryRepository.InsertCategory(ctx, book)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, book domain.Category) error {
	return s.categoryRepository.UpdateCategory(ctx, book)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int) error {
	return s.categoryRepository.DeleteCategory(ctx, id)
}
