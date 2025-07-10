package kubeforward

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sync"
)

var (
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: plugin.Namespace,
		Subsystem: "kubeforward",
		Name:      "request_duration_seconds",
		Help:      "Histogram of DNS request duration in kubeforward, in seconds",
		Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10),
	}, []string{"qtype", "rcode"})
)

var once sync.Once
