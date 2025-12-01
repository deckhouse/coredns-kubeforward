package kubeforward

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: plugin.Namespace,
		Subsystem: "kubeforward",
		Name:      "request_duration_seconds",
		Help:      "Histogram of DNS request duration in kubeforward, in seconds",
		Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10),
	}, []string{"qtype", "rcode"})

	SlowRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "kubeforward",
		Name:      "slow_requests_total",
		Help:      "Total number of DNS requests slower than the configured threshold",
	}, []string{"qtype", "rcode", "upstream"})
)
