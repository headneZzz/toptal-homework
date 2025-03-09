package middleware

import (
	"net/http"
	"strconv"
	"time"
	"toptal/internal/app/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.ActiveRequests.Inc()
		defer metrics.ActiveRequests.Dec()

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()
		handler := r.URL.Path

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
