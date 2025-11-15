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

	NXDomainByIPDomain = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "kubeforward",
		Name:      "nxdomain_by_ip_domain_total",
		Help:      "NXDOMAIN responses grouped by client IP and (normalized) domain",
	}, []string{"src_ip", "qname"})
)

var once sync.Once
