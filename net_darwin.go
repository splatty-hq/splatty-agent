//go:build darwin

package main

// Per-interface network counters on macOS require sysctl `net.link.generic.ifmibdata`
// with row indexing, which is significantly more work than the /proc/net/dev parser
// the Linux collector uses. Skipping for now — the darwin agent emits no network
// metrics.

type netCollector struct{}

func newNetCollector(_ config) collector { return &netCollector{} }

func (n *netCollector) collect() ([]metric, error) { return nil, nil }
