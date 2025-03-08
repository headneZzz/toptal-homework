package middleware

import (
	"net/http"
	"strconv"
	"time"
	"toptal/internal/app/metrics"
)

// MetricsMiddleware собирает метрики для HTTP запросов
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.ActiveRequests.Inc()
		defer metrics.ActiveRequests.Dec()

		// Создаем ResponseWriter, который может отслеживать статус ответа
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Обработка запроса
		next.ServeHTTP(rw, r)

		// Вычисляем длительность
		duration := time.Since(start).Seconds()

		// Получаем имя обработчика из пути
		handler := r.URL.Path

		// Записываем метрики
		metrics.RequestDuration.WithLabelValues(
			handler,
			r.Method,
			strconv.Itoa(rw.status),
		).Observe(duration)

		metrics.RequestsTotal.WithLabelValues(
			handler,
			r.Method,
			strconv.Itoa(rw.status),
		).Inc()

		if rw.status >= 400 {
			metrics.ErrorsTotal.WithLabelValues(
				handler,
				strconv.Itoa(rw.status),
			).Inc()
		}
	})
}

// responseWriter оборачивает http.ResponseWriter для отслеживания статуса ответа
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
