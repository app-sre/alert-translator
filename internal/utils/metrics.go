package utils

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	handledAlertsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alert_translator_handled_alerts",
			Help: "Increment for each alert handled at /alerts. Label for status of success/failure",
		},
		[]string{
			"status",
		},
	)
)

// register custom metrics at package import
func init() {
	prometheus.MustRegister(handledAlertsCounter)
	http.Handle("/metrics", promhttp.Handler())
}

func RecordMetrics(status string) {
	handledAlertsCounter.With(
		prometheus.Labels{
			"status": status,
		},
	).Inc()
}
