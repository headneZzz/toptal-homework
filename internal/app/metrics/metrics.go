package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestDuration измеряет длительность HTTP запросов
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests in seconds",
		},
		[]string{"handler", "method", "status"},
	)

	// RequestsTotal подсчитывает общее количество HTTP запросов
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"handler", "method", "status"},
	)

	// DatabaseQueryDuration измеряет длительность запросов к базе данных
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "database_query_duration_seconds",
			Help: "Duration of database queries in seconds",
		},
		[]string{"operation"},
	)

	// ActiveRequests отслеживает количество активных запросов
	ActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_requests",
			Help: "Number of active HTTP requests",
		},
	)

	// ErrorsTotal подсчитывает общее количество ошибок
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"handler", "code"},
	)
)
