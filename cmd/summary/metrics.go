package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpTotalRequests is the total number of http requests.
	httpTotalRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "http_requests_total",
			Namespace: appName,
			Help:      "Total number of http requests",
		},
		[]string{"path", "method", "status_code"},
	)

	// httpRequestDuration is the duration of the http request.
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "http_request_duration_seconds",
			Namespace: appName,
			Help:      "Duration of the http request",
		},
		[]string{"path", "method", "status_code"},
	)

	// httpRequestSize is the size of the http request.
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "http_request_size",
			Namespace: appName,
			Help:      "Size of the http request",
		},
		[]string{"path", "method", "status_code"},
	)
)
