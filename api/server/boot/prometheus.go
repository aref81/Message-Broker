package boot

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	successfulRPCCalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "broker",
			Name:      "successful_rpc_calls",
			Help:      "Number of successful gRPC calls",
		},
		[]string{"method"},
	)

	failedRPCCalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "broker",
			Name:      "failed_rpc_calls",
			Help:      "Number of failed gRPC calls",
		},
		[]string{"method"},
	)

	methodDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "broker",
			Name:      "server_method_duration_seconds",
			Help:      "Duration of gRPC request",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method"},
	)

	activeSubscribersGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "broker",
			Name:      "active_subscribers",
			Help:      "Number of active subscribers",
		})
)

func InitPrometheus() {
	reg := prometheus.NewRegistry()
	reg.MustRegister(successfulRPCCalls, failedRPCCalls, methodDuration, activeSubscribersGauge)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	http.Handle("/metrics", promHandler)
	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		logrus.Fatalf("Prometheus Failed To Start : %v", err)
		return
	}
}
