package handler

import (
	"context"
	"net/http"
	"time"
	"toptal/internal/app/handler/model"
)

const (
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
		Services:  make(map[string]model.Status),
	}

	dbStatus := s.checkDatabase(ctx)
	response.Services["database"] = dbStatus

	if dbStatus.Status == statusDown {
		response.Status = statusDown
	}

	if response.Status == statusDown {
		writeResponse(w, http.StatusServiceUnavailable, response)
	} else {
		writeResponseOK(w, response)
	}
}

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
