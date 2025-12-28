package engine

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// 请求总数
	enginRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "rule",
			Subsystem: "engine",
			Name:      "http_requests_total",
			Help:      "Total HTTP requests",
		},
		[]string{"name", "status"},
	)

	// 请求耗时
	enginRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "rule",
			Subsystem: "engine",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"name"},
	)
)

func init() {
	// 注册指标
	prometheus.MustRegister(enginRequestsTotal, enginRequestDuration)
}
