package service

import (
	"context"
	"toptal/internal/app/domain"
)

type CartService struct {
	cartRepository CartRepository
}

func NewCartService(repository CartRepository) *CartService {
	return &CartService{repository}
}

func (s *CartService) GetCart(ctx context.Context, userId int) ([]domain.Book, error) {
	return s.cartRepository.GetCart(ctx, userId)
}

func (s *CartService) AddToCart(ctx context.Context, userId int, bookId int) error {
	return s.cartRepository.AddToCart(ctx, userId, bookId)
}

func (s *CartService) RemoveFromCart(ctx context.Context, userId int, bookId int) error {
	return s.cartRepository.RemoveFromCart(ctx, userId, bookId)
}

func (s *CartService) Purchase(ctx context.Context, userId int) error {
	return s.cartRepository.Purchase(ctx, userId)
}
