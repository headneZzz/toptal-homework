package service

import (
	"context"
	"fmt"
	"log/slog"
	"toptal/internal/app/domain"
)

type BookService struct {
	bookRepository BookRepository
	authService    AuthService
}

func NewBookService(bookRepository BookRepository, authService AuthService) *BookService {
	return &BookService{bookRepository, authService}
}

func (s *BookService) GetBookById(ctx context.Context, id int) (domain.Book, error) {
	return s.bookRepository.GetById(ctx, id)
}

func (s *BookService) GetAvailableBooks(ctx context.Context, categoryIds []int, limit, offset int) ([]domain.Book, error) {
	return s.bookRepository.GetByCategories(ctx, categoryIds, limit, offset)
}

func (s *BookService) CreateBook(ctx context.Context, book domain.Book) error {
	slog.Info("BookService.CreateBook", "request", book)
	if err := s.authService.checkAdmin(ctx); err != nil {
		return fmt.Errorf("check admin failed: %w", err)
	}
	return s.bookRepository.Create(ctx, book)
}

func (s *BookService) UpdateBook(ctx context.Context, book domain.Book) error {
	if err := s.authService.checkAdmin(ctx); err != nil {
		return fmt.Errorf("check admin failed: %w", err)
	}
	return s.bookRepository.Update(ctx, book)
}

func (s *BookService) DeleteBook(ctx context.Context, id int) error {
	if err := s.authService.checkAdmin(ctx); err != nil {
		return fmt.Errorf("check admin failed: %w", err)
	}
	return s.bookRepository.Delete(ctx, id)
}
