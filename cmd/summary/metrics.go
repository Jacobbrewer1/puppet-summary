package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/gorilla/mux"
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

func metricsWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var path string
		route := mux.CurrentRoute(r)
		if route != nil { // The route may be nil if the request is not routed.
			var err error
			path, err = route.GetPathTemplate()
			if err != nil {
				// An error here is only returned if the route does not define a path.
				slog.Error("Error getting path template", slog.String(logging.KeyError, err.Error()))
				path = r.URL.Path // If the route does not define a path, use the URL path.
			}
		} else {
			path = r.URL.Path // If the route is nil, use the URL path.
		}

		reqSize := r.ContentLength

		cw := request.NewClientWriter(w)
		h.ServeHTTP(cw, r)

		httpTotalRequests.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Inc()
		httpRequestDuration.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Observe(time.Since(start).Seconds())
		httpRequestSize.WithLabelValues(path, r.Method, fmt.Sprintf("%d", cw.StatusCode())).Observe(float64(reqSize))
	}
}
