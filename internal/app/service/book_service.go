package service

import (
	"context"
	"log/slog"
	"toptal/internal/app/domain"
)

type BookService struct {
	bookRepository BookRepository
	userService    UserService
}

func NewBookService(bookRepository BookRepository, userService UserService) *BookService {
	return &BookService{bookRepository, userService}
}

func (s *BookService) GetBookById(ctx context.Context, id int) (domain.Book, error) {
	return s.bookRepository.GetById(ctx, id)
}

func (s *BookService) GetAvailableBooks(ctx context.Context, categoryIds []int) ([]domain.Book, error) {
	return s.bookRepository.GetByCategories(ctx, categoryIds)
}

func (s *BookService) CreateBook(ctx context.Context, book domain.Book) error {
	slog.Info("BookService.CreateBook", "request", book)
	if err := s.userService.checkAdmin(ctx); err != nil {
		return err
	}
	return s.bookRepository.Create(ctx, book)
}

func (s *BookService) UpdateBook(ctx context.Context, book domain.Book) error {
	if err := s.userService.checkAdmin(ctx); err != nil {
		return err
	}
	return s.bookRepository.Update(ctx, book)
}

func (s *BookService) DeleteBook(ctx context.Context, id int) error {
	if err := s.userService.checkAdmin(ctx); err != nil {
		return err
	}
	return s.bookRepository.Delete(ctx, id)
}
