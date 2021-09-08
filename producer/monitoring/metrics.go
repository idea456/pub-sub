package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Declare all metrics here

var TotalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of incoming requests",
	},
	[]string{"path", "method"})

var ResponseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "HTTP response status codes",
	},
	[]string{"status", "path", "method"})

var Latency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_response_time_seconds",
		Help: "Latency of HTTP requests",
	}, []string{"path", "method"})
