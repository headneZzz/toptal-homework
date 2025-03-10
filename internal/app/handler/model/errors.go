package model

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ProblemDetail struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

func WriteProblemDetail(w http.ResponseWriter, status int, title, detail, instance string) {
	pd := ProblemDetail{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(pd.Status)
	if err := json.NewEncoder(w).Encode(pd); err != nil {
		slog.Error("Failed to encode problem detail", "error", err)
		http.Error(w, "Failed to encode problem detail", http.StatusInternalServerError)
	}
}

func InternalServerError(w http.ResponseWriter, instance string) {
	WriteProblemDetail(w, http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", instance)
}

func NotFound(w http.ResponseWriter, detail, instance string) {
	WriteProblemDetail(w, http.StatusNotFound, "Not Found", detail, instance)
}

func ValidationError(w http.ResponseWriter, detail, instance string) {
	WriteProblemDetail(w, http.StatusBadRequest, "Validation Error", detail, instance)
}

func AlreadyExists(w http.ResponseWriter, detail, instance string) {
	WriteProblemDetail(w, http.StatusConflict, "Already Exists", detail, instance)
}

func InvalidRequest(w http.ResponseWriter, detail, instance string) {
	WriteProblemDetail(w, http.StatusBadRequest, "Invalid Request", detail, instance)
}

func Unauthorized(w http.ResponseWriter, detail, instance string) {
	WriteProblemDetail(w, http.StatusUnauthorized, "Unauthorized", detail, instance)
}

func Forbidden(w http.ResponseWriter, detail, instance string) {
	WriteProblemDetail(w, http.StatusForbidden, "Forbidden", detail, instance)
}
