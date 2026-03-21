package kubeforward

import (
	"fmt"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/proxy"
	"github.com/miekg/dns"
)

type KubeForwardConfig struct {
	Namespace           string
	ServiceName         string
	PortName            string
	Expire              time.Duration
	UpstreamReadTimeout time.Duration
	SlowThreshold       time.Duration
	SlowLogEnabled      bool
	opts                proxy.Options
}

// ParseConfig parse conf CoreFile
func ParseConfig(c *caddy.Controller) (*KubeForwardConfig, error) {
	config := &KubeForwardConfig{
		Expire:              10 * time.Second,  // Default value
		UpstreamReadTimeout: 300 * time.Second, // Default value
		SlowThreshold:       0,
		SlowLogEnabled:      false,
		opts: proxy.Options{
			ForceTCP:           false,
			PreferUDP:          false,
			HCRecursionDesired: true,
			HCDomain:           ".",
		},
	}

	c.RemainingArgs()
	// Checking the presence of a parameter block
	for c.NextBlock() {
		switch c.Val() {
		case "namespace":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			config.Namespace = c.Val()
		case "service_name":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			config.ServiceName = c.Val()
		case "port_name":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			config.PortName = c.Val()
		case "expire":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			duration, err := time.ParseDuration(c.Val())
			if err != nil {
				return nil, fmt.Errorf("invalid expire duration: %v", err)
			}
			config.Expire = duration
		case "health_check":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			for {
				val := c.Val()
				switch val {
				case "no_rec":
					config.opts.HCRecursionDesired = false
				case "domain":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					hcDomain := c.Val()
					if _, ok := dns.IsDomainName(hcDomain); !ok {
						return nil, fmt.Errorf("health_check: invalid domain name %s", hcDomain)
					}
					config.opts.HCDomain = plugin.Name(hcDomain).Normalize()
				default:
					if _, err := time.ParseDuration(val); err == nil {
						return nil, fmt.Errorf("health_check duration is not supported by kubeforward; use upstream_read_timeout")
					}
					return nil, fmt.Errorf("health_check: unknown option %s", val)
				}
				if !c.NextArg() {
					break
				}
			}

		case "upstream_read_timeout":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			duration, err := time.ParseDuration(c.Val())
			if err != nil {
				return nil, fmt.Errorf("invalid upstream_read_timeout duration: %v", err)
			}
			config.UpstreamReadTimeout = duration
		case "slow_threshold":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			duration, err := time.ParseDuration(c.Val())
			if err != nil {
				return nil, fmt.Errorf("invalid slow_threshold duration: %v", err)
			}
			config.SlowThreshold = duration
		case "slow_log":
			if c.NextArg() {
				return nil, c.ArgErr()
			}
			config.SlowLogEnabled = true
		case "force_tcp":
			config.opts.ForceTCP = true
		case "prefer_udp":
			config.opts.PreferUDP = true

		default:
			return nil, c.Errf("unknown parameter: %s", c.Val())
		}
	}

	// Checking the required parameters
	if config.Namespace == "" || config.ServiceName == "" || config.PortName == "" {
		return nil, fmt.Errorf("namespace, servicename, and portname are required parameters")
	}

	return config, nil
}
