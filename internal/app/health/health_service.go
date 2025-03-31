package health

import (
	"context"
	"fmt"

	"toptal/internal/pkg/pg"
)

type Service struct {
	db *pg.DB
}

func NewHealthService(db *pg.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CheckDatabase(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	var result int
	if err := s.db.QueryRowxContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected database response: %d", result)
	}

	return nil
}
