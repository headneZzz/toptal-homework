package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"toptal/internal/app/handler/model"
)

const (
	// Версия API
	apiVersion = "1.0.0"

	// Статусы здоровья
	statusUp   = "UP"
	statusDown = "DOWN"
)

// @Summary Health check
// @Description Get the health status of the server and its dependencies
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} model.HealthResponse
// @Failure 503 {object} model.HealthResponse
// @Router /health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := model.HealthResponse{
		Status:    statusUp,
		Timestamp: time.Now(),
		Version:   apiVersion,
		Services:  make(map[string]model.Status),
	}

	// Проверка базы данных
	dbStatus := s.checkDatabase(ctx)
	response.Services["database"] = dbStatus

	// Если какой-либо из сервисов не работает, меняем общий статус
	if dbStatus.Status == statusDown {
		response.Status = statusDown
	}

	w.Header().Set("Content-Type", "application/json")
	if response.Status == statusDown {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// checkDatabase проверяет доступность базы данных
func (s *Server) checkDatabase(ctx context.Context) model.Status {
	err := s.healthService.CheckDatabase(ctx)
	if err != nil {
		return model.Status{
			Status:  statusDown,
			Message: "database check failed: " + err.Error(),
		}
	}

	return model.Status{
		Status:  statusUp,
		Message: "database is healthy",
	}
}
