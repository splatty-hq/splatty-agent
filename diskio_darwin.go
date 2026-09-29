//go:build darwin

package main

// Per-disk I/O counters on macOS live behind IOKit's IOBlockStorageDriver
// statistics, which need cgo. The darwin agent emits no disk I/O metrics.

type diskIOCollector struct{}

func newDiskIOCollector(_ config) collector { return &diskIOCollector{} }

func (d *diskIOCollector) collect() ([]metric, error) { return nil, nil }
