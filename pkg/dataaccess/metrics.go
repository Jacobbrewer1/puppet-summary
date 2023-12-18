package dataaccess

import "github.com/prometheus/client_golang/prometheus"

// DatabaseLatency is the duration of database queries.
var DatabaseLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "database_latency",
		Help: "Duration of database queries",
	},
	[]string{"query"},
)
