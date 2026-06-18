//go:build darwin

package main

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

type memCollector struct{}

func newMemCollector(_ config) collector { return &memCollector{} }

func (m *memCollector) collect() ([]metric, error) {
	var out []metric

	if data, err := unix.SysctlRaw("hw.memsize"); err == nil && len(data) >= 8 {
		total := binary.LittleEndian.Uint64(data)
		out = append(out, metric{Name: "mem.total_bytes", Value: float64(total)})
	}

	// struct xsw_usage { u_int64_t total; u_int64_t avail; u_int64_t used;
	//                    u_int32_t pagesize; boolean_t encrypted; }
	// The first 24 bytes (total/avail/used) are all we need.
	if data, err := unix.SysctlRaw("vm.swapusage"); err == nil && len(data) >= 24 {
		total := binary.LittleEndian.Uint64(data[0:8])
		avail := binary.LittleEndian.Uint64(data[8:16])
		used := binary.LittleEndian.Uint64(data[16:24])
		out = append(out,
			metric{Name: "swap.total_bytes", Value: float64(total)},
			metric{Name: "swap.free_bytes", Value: float64(avail)},
			metric{Name: "swap.used_bytes", Value: float64(used)},
		)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no memory metrics available via sysctl")
	}
	return out, nil
}
