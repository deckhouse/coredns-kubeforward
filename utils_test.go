package kubeforward

import (
	"strings"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/proxy"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      KubeForwardConfig
		expectErr     bool
		expectedError string
	}{
		{
			name: "Valid config with all supported parameters",
			input: `kubeforward {
				namespace kube-system
				service_name d8-kube-dns
				port_name dns
				expire 10m
				upstream_read_timeout 5s
				health_check no_rec domain example.org
				prefer_udp
				slow_threshold 200ms
				slow_log
			}`,
			expected: KubeForwardConfig{
				Namespace:           "kube-system",
				ServiceName:         "d8-kube-dns",
				PortName:            "dns",
				Expire:              10 * time.Minute,
				UpstreamReadTimeout: 5 * time.Second,
				SlowThreshold:       200 * time.Millisecond,
				SlowLogEnabled:      true,
				opts: proxy.Options{
					PreferUDP:          true,
					HCRecursionDesired: false,
					HCDomain:           "example.org.",
				},
			},
			expectErr: false,
		},
		{
			name: "Config missing namespace",
			input: `kubeforward {
				service_name d8-kube-dns
				port_name dns
				upstream_read_timeout 5s
			}`,
			expectErr:     true,
			expectedError: "namespace, servicename, and portname are required parameters",
		},
		{
			name: "Config with invalid expire value",
			input: `kubeforward {
				namespace kube-system
				service_name d8-kube-dns
				port_name dns
				expire not-a-duration
			}`,
			expectErr:     true,
			expectedError: "invalid expire duration",
		},
		{
			name: "Config with missing health_check value",
			input: `kubeforward {
				namespace kube-system
				service_name d8-kube-dns
				port_name dns
				health_check
			}`,
			expectErr:     true,
			expectedError: "Wrong argument count or unexpected line ending after 'health_check'",
		},
		{
			name: "Config with unsupported health_check duration",
			input: `kubeforward {
				namespace kube-system
				service_name d8-kube-dns
				port_name dns
				health_check 500ms
			}`,
			expectErr:     true,
			expectedError: "health_check duration is not supported by kubeforward; use upstream_read_timeout",
		},
		{
			name: "Config with invalid health_check domain",
			input: `kubeforward {
				namespace kube-system
				service_name d8-kube-dns
				port_name dns
				health_check domain example..org
			}`,
			expectErr:     true,
			expectedError: "health_check: invalid domain name",
		},
		{
			name: "Minimal valid config",
			input: `kubeforward {
				namespace kube-system
				service_name d8-kube-dns
				port_name dns
			}`,
			expected: KubeForwardConfig{
				Namespace:           "kube-system",
				ServiceName:         "d8-kube-dns",
				PortName:            "dns",
				Expire:              10 * time.Second,
				UpstreamReadTimeout: 300 * time.Second,
				SlowThreshold:       0,
				SlowLogEnabled:      false,
				opts: proxy.Options{
					HCRecursionDesired: true,
					HCDomain:           ".",
				},
			},
			expectErr: false,
		},
		{
			name: "Config with force_tcp",
			input: `kubeforward {
				namespace kube-system
				service_name d8-kube-dns
				port_name dns
				force_tcp
			}`,
			expected: KubeForwardConfig{
				Namespace:           "kube-system",
				ServiceName:         "d8-kube-dns",
				PortName:            "dns",
				Expire:              10 * time.Second,
				UpstreamReadTimeout: 300 * time.Second,
				opts: proxy.Options{
					ForceTCP:           true,
					HCRecursionDesired: true,
					HCDomain:           ".",
				},
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := caddy.NewTestController("dns", test.input)

			config, err := ParseConfig(controller)

			if test.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), test.expectedError) {
					t.Fatalf("expected error containing %q, got %q", test.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.Namespace != test.expected.Namespace {
				t.Errorf("expected namespace %q, got %q", test.expected.Namespace, config.Namespace)
			}
			if config.ServiceName != test.expected.ServiceName {
				t.Errorf("expected service_name %q, got %q", test.expected.ServiceName, config.ServiceName)
			}
			if config.PortName != test.expected.PortName {
				t.Errorf("expected port_name %q, got %q", test.expected.PortName, config.PortName)
			}
			if config.Expire != test.expected.Expire {
				t.Errorf("expected expire %v, got %v", test.expected.Expire, config.Expire)
			}
			if config.UpstreamReadTimeout != test.expected.UpstreamReadTimeout {
				t.Errorf("expected upstream_read_timeout %v, got %v", test.expected.UpstreamReadTimeout, config.UpstreamReadTimeout)
			}
			if config.SlowThreshold != test.expected.SlowThreshold {
				t.Errorf("expected slow_threshold %v, got %v", test.expected.SlowThreshold, config.SlowThreshold)
			}
			if config.SlowLogEnabled != test.expected.SlowLogEnabled {
				t.Errorf("expected slow_log %v, got %v", test.expected.SlowLogEnabled, config.SlowLogEnabled)
			}
			if config.opts.HCRecursionDesired != test.expected.opts.HCRecursionDesired {
				t.Errorf("expected health_check recursion_desired %v, got %v", test.expected.opts.HCRecursionDesired, config.opts.HCRecursionDesired)
			}
			if config.opts.HCDomain != test.expected.opts.HCDomain {
				t.Errorf("expected health_check domain %q, got %q", test.expected.opts.HCDomain, config.opts.HCDomain)
			}
			if config.opts.ForceTCP != test.expected.opts.ForceTCP {
				t.Errorf("expected force_tcp %v, got %v", test.expected.opts.ForceTCP, config.opts.ForceTCP)
			}
			if config.opts.PreferUDP != test.expected.opts.PreferUDP {
				t.Errorf("expected prefer_udp %v, got %v", test.expected.opts.PreferUDP, config.opts.PreferUDP)
			}
		})
	}
}
