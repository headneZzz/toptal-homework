package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func writeResponse(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func writeResponseOK(w http.ResponseWriter, response any) {
	writeResponse(w, http.StatusOK, response)
}

func writeResponseCreated(w http.ResponseWriter, response interface{}) {
	writeResponse(w, http.StatusCreated, response)
}
