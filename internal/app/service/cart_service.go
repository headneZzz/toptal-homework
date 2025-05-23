package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"toptal/internal/app/config"
	"toptal/internal/app/domain"
)

type CartService struct {
	cartRepository CartRepository
	config         *config.CartConfig
}

func NewCartService(repository CartRepository, cfg *config.CartConfig) *CartService {
	return &CartService{cartRepository: repository, config: cfg}
}

func (s *CartService) GetCart(ctx context.Context, userId int) ([]domain.Book, error) {
	return s.cartRepository.GetCart(ctx, userId)
}

func (s *CartService) AddToCart(ctx context.Context, userId, bookId int) error {
	if err := s.cartRepository.AddToCart(ctx, userId, bookId); err != nil {
		slog.Error("failed to add to cart", "error", err)
		return fmt.Errorf("failed to add book to cart: %w", err)
	}
	return nil
}

func (s *CartService) RemoveFromCart(ctx context.Context, userId, bookId int) error {
	return s.cartRepository.RemoveFromCart(ctx, userId, bookId)
}

func (s *CartService) Purchase(ctx context.Context, userId int) error {
	return s.cartRepository.Purchase(ctx, userId)
}

func (s *CartService) StartCartCleanerJob(ctx context.Context) {
	ticker := time.NewTicker(s.config.CleanupInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Info("Cleaning Carts")
				if err := s.cartRepository.CleanExpiredCarts(ctx); err != nil {
					slog.Error(err.Error())
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	slog.Info("Cart cleaner job started", "interval minutes", s.config.CleanupInterval.Minutes())
}
