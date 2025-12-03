package kubeforward

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/metadata"
	"github.com/coredns/coredns/plugin/pkg/proxy"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/miekg/dns"
)

// KubeForward main struct of plugin
type KubeForward struct {
	Next           plugin.Handler
	Namespace      string
	ServiceName    string
	forwardTo      []string
	forwarder      *forward.Forward
	options        proxy.Options
	cond           *sync.Cond
	slowThreshold  time.Duration
	slowLogEnabled bool
}

func (df *KubeForward) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	df.cond.L.Lock()
	if df.forwarder == nil {
		df.cond.Wait()
	}
	forwarder := df.forwarder
	df.cond.L.Unlock()

	rec := &responseRecorder{ResponseWriter: w}
	start := time.Now()
	rcode, err := forwarder.ServeDNS(ctx, rec, r)
	elapsed := time.Since(start)
	rcodeStr := dns.RcodeToString[rec.recordedRcode(rcode)]

	df.observeRequest(ctx, r, rcodeStr, elapsed)

	return rcode, err
}

func (df *KubeForward) observeRequest(ctx context.Context, r *dns.Msg, rcodeStr string, elapsed time.Duration) {
	if len(r.Question) == 0 {
		return
	}

	q := r.Question[0]
	qtype := dns.TypeToString[q.Qtype]

	RequestDuration.WithLabelValues(qtype, rcodeStr).Observe(elapsed.Seconds())

	if df.slowThreshold > 0 && elapsed > df.slowThreshold {
		upstream := upstreamFromContext(ctx)
		SlowRequests.WithLabelValues(qtype, rcodeStr, upstream).Inc()
		if df.slowLogEnabled {
			log.Printf("[kubeforward] slow query %s %s took %v (rcode=%s, upstream=%s)",
				qtype, q.Name, elapsed, rcodeStr, upstream)
		}
	}
}

// need `metadata` plugin before kubeforward in chain
func upstreamFromContext(ctx context.Context) string {
	if upstreamFunc := metadata.ValueFunc(ctx, "forward/upstream"); upstreamFunc != nil {
		if upstream := upstreamFunc(); upstream != "" {
			return upstream
		}
	}

	return "unknown"
}

type responseRecorder struct {
	dns.ResponseWriter
	rcode int
}

func (r *responseRecorder) WriteMsg(res *dns.Msg) error {
	r.rcode = res.Rcode
	return r.ResponseWriter.WriteMsg(res)
}

func (r *responseRecorder) recordedRcode(defaultRcode int) int {
	if r.rcode != 0 {
		return r.rcode
	}
	return defaultRcode
}

// UpdateForwardServers update list servers for forward requests
func (df *KubeForward) UpdateForwardServers(newServers []string, config KubeForwardConfig) {
	df.cond.L.Lock()

	newForwarder := forward.New()

	for _, server := range newServers {
		proxyInstance := proxy.NewProxy(server, server, transport.DNS)
		proxyInstance.SetExpire(config.Expire)
		proxyInstance.SetReadTimeout(config.HealthCheck)
		newForwarder.SetProxy(proxyInstance)
		newForwarder.SetProxyOptions(df.options)
	}

	oldForwarder := df.forwarder

	// Fill up list servers
	df.forwarder = newForwarder
	df.forwardTo = newServers
	df.cond.Broadcast()
	df.cond.L.Unlock()

	if oldForwarder != nil {
		for _, oldProxy := range oldForwarder.List() {
			oldProxy.Stop()
		}
	}

	log.Printf("[kubeforward] Forward servers updated: %v", newServers)
}

// Name return plugin name
func (df *KubeForward) Name() string { return "kubeforward" }
