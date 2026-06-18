//go:build darwin

package main

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

type loadCollector struct{}

func newLoadCollector(_ config) collector { return &loadCollector{} }

func (l *loadCollector) collect() ([]metric, error) {
	data, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		return nil, fmt.Errorf("sysctl vm.loadavg: %w", err)
	}
	// struct loadavg { fixpt_t ldavg[3]; long fscale; }
	// fixpt_t = uint32 (4 bytes), long = int64 on 64-bit darwin. The struct
	// pads to 24 bytes (12 ldavg + 4 padding + 8 fscale).
	if len(data) < 24 {
		return nil, fmt.Errorf("vm.loadavg returned %d bytes (expected >= 24)", len(data))
	}
	ldavg := [3]uint32{
		binary.LittleEndian.Uint32(data[0:4]),
		binary.LittleEndian.Uint32(data[4:8]),
		binary.LittleEndian.Uint32(data[8:12]),
	}
	fscale := binary.LittleEndian.Uint64(data[16:24])
	if fscale == 0 {
		return nil, fmt.Errorf("vm.loadavg fscale is zero")
	}
	f := float64(fscale)
	return []metric{
		{Name: "load.1", Value: float64(ldavg[0]) / f},
		{Name: "load.5", Value: float64(ldavg[1]) / f},
		{Name: "load.15", Value: float64(ldavg[2]) / f},
	}, nil
}
