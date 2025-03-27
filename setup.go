package kubeforward

import (
	"context"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"log"
	"sync"
)

func init() { plugin.Register("kubeforward", setup) }

func setup(c *caddy.Controller) error {

	version := "0.3.5"

	log.Printf("\033[34m[kubeforward] version: %s\033[0m\n", version)

	// parse config
	config, err := ParseConfig(c)
	if err != nil {
		return err
	}

	kubeForwardPlugin := &KubeForward{
		Namespace:   config.Namespace,
		ServiceName: config.ServiceName, //kubernetes.io/service-name=d8-kube-dns
		forwarder:   nil,
		options:     config.opts,
		cond:        sync.NewCond(&sync.Mutex{}),
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		kubeForwardPlugin.Next = next
		return kubeForwardPlugin
	})

	// Context for properly shutdown goroutine
	ctx, cancel := context.WithCancel(context.Background())

	c.OnStartup(func() error {
		log.Printf("[kubeforward] Starting with namespace=%s, service_name=%s\n", config.Namespace, config.ServiceName)
		// Start go routine for watch EndpointSlice
		go func() {
			err := startEndpointSliceWatcher(ctx, config.Namespace, config.ServiceName, config.PortName, func(newServers []string) {
				kubeForwardPlugin.UpdateForwardServers(newServers, *config)
				log.Printf("[kubeforward] Updated servers namespace%s, service_name=%s\n: %v", config.Namespace, config.ServiceName, newServers)
			})

			if err != nil {
				log.Printf("[kubeforward] Error starting EndpointSlice watcher with label kubernetes.io/service-name=%s: %v", config.ServiceName, err)
			}
		}()

		return nil
	})

	c.OnShutdown(func() error {
		log.Printf("[kubeforward] Shutting down with namespace=%s\n", config.Namespace)
		cancel()
		return nil
	})

	return nil
}
