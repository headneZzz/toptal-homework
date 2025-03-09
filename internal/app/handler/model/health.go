package model

import "time"

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]Status `json:"services"`
}

type Status struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
