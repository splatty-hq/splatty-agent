//go:build darwin

package main

import (
	"encoding/binary"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

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

	pageSize := uint64(4096)
	if v, err := unix.SysctlUint32("hw.pagesize"); err == nil && v > 0 {
		pageSize = uint64(v)
	}

	if vmStatOut, err := exec.Command("/usr/bin/vm_stat").Output(); err == nil {
		pages := parseVMStat(string(vmStatOut))
		used := (pages["Pages active"] + pages["Pages wired down"] + pages["Pages occupied by compressor"]) * pageSize
		available := (pages["Pages free"] + pages["Pages inactive"] + pages["Pages speculative"]) * pageSize
		if used > 0 {
			out = append(out, metric{Name: "mem.used_bytes", Value: float64(used)})
		}
		if available > 0 {
			out = append(out, metric{Name: "mem.available_bytes", Value: float64(available)})
		}
	}

	// struct xsw_usage { u_int64_t total; u_int64_t avail; u_int64_t used; ... }
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
		return nil, fmt.Errorf("no memory metrics available")
	}
	return out, nil
}

func parseVMStat(s string) map[string]uint64 {
	out := map[string]uint64{}
	for _, line := range strings.Split(s, "\n") {
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		val = strings.TrimSuffix(val, ".")
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			continue
		}
		out[key] = n
	}
	return out
}
