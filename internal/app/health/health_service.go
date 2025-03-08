package health

import (
	"context"
	"fmt"

	"toptal/internal/pkg/pg"
)

type HealthService struct {
	db *pg.DB
}

func NewHealthService(db *pg.DB) *HealthService {
	return &HealthService{db: db}
}

// CheckDatabase проверяет доступность базы данных
func (s *HealthService) CheckDatabase(ctx context.Context) error {
	// Проверяем соединение с БД
	if err := s.db.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Проверяем, что можем выполнить простой запрос
	var result int
	if err := s.db.DB.QueryRowxContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected database response: %d", result)
	}

	return nil
}
