# kubeforward
CoreDNS external plugin for node-local-dns

**kubeforward** is a [CoreDNS](https://coredns.io/) plugin designed to dynamically monitor changes in Kubernetes Services and automatically update the list of DNS forwarders. It observes `EndpointSlices` associated with a specified Service, dynamically adjusting the DNS forwarding configuration as endpoints are added, removed, or updated.

## Key Features

- **Dynamic DNS Forwarding Updates**: Automatically tracks changes in `EndpointSlices`, ensuring the DNS forwarding configuration remains current without manual intervention.
- **Enhanced Reliability and Fault Tolerance**: In the event of an API server failure, CoreDNS continues to use the last known list of endpoints, maintaining DNS service stability.

## Use Case

The **kubeforward** plugin is particularly useful for organizing a node-local-dns caching mechanism. In the standard scheme, `ClusterIP` is used as the upstream, delegating load balancing responsibility to the CNI. However, in the event of an API server failure, the CNI may not be aware of which upstream endpoints are alive. **kubeforward** allows CoreDNS to handle upstream health checks independently and, in the case of an API server failure, it will still retain the list of endpoints and load balance the requests.

## Installation

1. **Add the Plugin to CoreDNS**:
   - Clone the CoreDNS repository:
     ```bash
     git clone https://github.com/coredns/coredns
     ```
   - Navigate to the project directory:
     ```bash
     cd coredns
     ```
   - Add the **kubeforward** plugin to the `plugin.cfg` file:
     ```text
     kubeforward:github.com/deckhouse/coredns-kubeforward
     ```
     Ensure that this line is added before the `forward:forward` line to maintain the correct order of plugin execution.

2. **Build CoreDNS with the New Plugin**:
   - Execute the following commands:
     ```bash
     go get github.com/deckhouse/coredns-kubeforward
     go generate
     go build
     ```
     This will generate and build CoreDNS with the **kubeforward** plugin included.

## Configuration

The plugin is configured in the `Corefile` as follows:

```coredns
.:53 {
    errors
    log
    kubeforward {
        namespace kube-system
        service_name kube-dns
        port_name dns
        expire 10m
        health_check 5s
        prefer_udp
        force_tcp
        slow_threshold 300ms
        slow_log
    }
}
```

## Configuration Parameters

- `namespace` (required): Specifies the Kubernetes namespace where the target Service resides.

- `service_name` (required): The name of the Service to which DNS queries will be forwarded.

- `port_name`: The name of the port in the Service resource responsible for handling DNS queries.

- `expire`: Time after which cached connections expire. Default is 10s.

- `health_check`: Interval for health checking of upstream servers. Default is 300s.

- `force_tcp`: Forces the use of TCP for forwarding queries.

- `prefer_udp`: Prefers the use of UDP for forwarding queries.

- `slow_threshold`: Duration threshold; DNS queries handled by `kubeforward` that take longer than this value are counted in `slow_requests_total`. Set to `0` (default) to disable slow counting.

- `slow_log`: When present, logs slow queries (those over `slow_threshold`) to stdout. The metric `slow_requests_total` is emitted regardless of this flag.

## Metrics

- `coredns_kubeforward_request_duration_seconds{qtype,rcode}`: Histogram of request durations.
- `coredns_kubeforward_slow_requests_total{qtype,rcode,upstream}`: Counter of requests slower than `slow_threshold`. To populate the `upstream` label, include the `metadata` plugin before `kubeforward` in the Corefile.

## Limitations

Limited Support for Forward Plugin Options: The plugin utilizes the functionality of the forward plugin for serving DNS under the hood but does not support the full list of classic forward options due to the lack of a public interface for configuring options.

## License

This project is distributed under the Apache License Version 2.0. See the LICENSE file for details.
