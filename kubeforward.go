package kubeforward

import (
	"context"
	"log"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/pkg/proxy"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/miekg/dns"
)

// KubeForward main struct of plugin
type KubeForward struct {
	Next        plugin.Handler
	Namespace   string
	ServiceName string
	forwardTo   []string
	mu          sync.RWMutex
	forwarder   *forward.Forward
	options     proxy.Options
	cond        *sync.Cond
}

func (df *KubeForward) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	//wait forwarder
	df.cond.L.Lock()

	if df.forwarder == nil {
		df.cond.Wait()
	}

	df.cond.L.Unlock()

	return df.forwarder.ServeDNS(ctx, w, r)
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

	log.Printf("[dynamicforward] Forward servers updated: %v", newServers)
}

// Name return plugin name
func (df *KubeForward) Name() string { return "dynamicforward" }
