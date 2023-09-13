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
			Name:      "grpc_server_successful_calls",
			Help:      "Number of successful gRPC calls by method.",
		},
		[]string{"method"},
	)
	//
	//failedRPCCalls = prometheus.NewCounterVec(
	//	prometheus.CounterOpts{
	//		Namespace: "broker",
	//		Name: "grpc_server_failed_calls",
	//		Help: "Number of failed gRPC calls by method.",
	//	},
	//	[]string{"method"},
	//)
	//
	//methodDuration = prometheus.NewHistogramVec(
	//	prometheus.HistogramOpts{
	//		Namespace: "broker",
	//		Name:    "grpc_server_method_duration_seconds",
	//		Help:    "Duration of gRPC calls by method.",
	//		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	//	},
	//	[]string{"method"},
	//)
	//
	//activeSubscribersGauge = prometheus.NewGauge(
	//	prometheus.GaugeOpts{
	//		Namespace: "broker",
	//		Name: "active_subscribers",
	//		Help: "Number of active subscribers",
	//	})
)

func InitPrometheus() {
	reg := prometheus.NewRegistry()
	reg.MustRegister(successfulRPCCalls)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	//// Create a router for Prometheus metrics
	//router := mux.NewRouter()
	//router.Handle("/metrics", promHandler)

	http.Handle("/metrics", promHandler)
	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		logrus.Fatalf("Prometheus Failed To Start : %v", err)
		return
	}
}
