//go:build darwin

package main

// CPU usage percent on macOS requires the Mach host_statistics() call
// (HOST_CPU_LOAD_INFO), which is not reachable from Go's stdlib without cgo.
// We emit nothing here; CPU charts will be empty on darwin agents until/unless
// we add a cgo-backed collector.

type cpuCollector struct{}

func newCPUCollector(_ config) collector { return &cpuCollector{} }

func (c *cpuCollector) collect() ([]metric, error) { return nil, nil }
