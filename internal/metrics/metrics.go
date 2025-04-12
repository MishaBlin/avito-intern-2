package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pvz_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	ResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pvz_http_response_time_seconds",
			Help:    "HTTP response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	PvzCreatedCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total number of created pickup points",
		},
	)

	ReceptionCreatedCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "reception_created_total",
			Help: "Total number of created order receptions",
		},
	)

	ProductAddedCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "product_added_total",
			Help: "Total number of added products",
		},
	)
)
