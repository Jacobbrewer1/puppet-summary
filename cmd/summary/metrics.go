package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpTotalRequests is the total number of http requests.
	httpTotalRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_http_total_requests", appName),
			Help: "Total number of http requests",
		},
		[]string{"path", "method", "status_code"},
	)

	// httpRequestDuration is the duration of the http request.
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: fmt.Sprintf("%s_http_request_duration", appName),
			Help: "Duration of the http request",
		},
		[]string{"path", "method", "status_code"},
	)
)
