package middleware

import (
	"avito-intern/internal/metrics"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = "unknown"
		}

		metrics.RequestCount.WithLabelValues(
			r.Method,
			routePattern,
			strconv.Itoa(ww.Status()),
		).Inc()

		metrics.ResponseTime.WithLabelValues(
			r.Method,
			routePattern,
		).Observe(time.Since(start).Seconds())
	})
}
