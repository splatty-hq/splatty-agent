//go:build linux

package main

import (
	"os"
	"path/filepath"
)

type memCollector struct {
	procRoot string
}

func newMemCollector(cfg config) collector {
	return &memCollector{procRoot: cfg.ProcRoot}
}

func (m *memCollector) collect() ([]metric, error) {
	f, err := os.Open(filepath.Join(m.procRoot, "meminfo"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info := parseMeminfo(f)
	var out []metric
	if v, ok := info["MemTotal"]; ok {
		out = append(out, metric{Name: "mem.total_bytes", Value: float64(v)})
	}
	if v, ok := info["MemAvailable"]; ok {
		out = append(out, metric{Name: "mem.available_bytes", Value: float64(v)})
	}
	if total, ok := info["MemTotal"]; ok {
		if avail, ok2 := info["MemAvailable"]; ok2 {
			out = append(out, metric{Name: "mem.used_bytes", Value: float64(total - avail)})
		}
	}
	if v, ok := info["SwapTotal"]; ok {
		out = append(out, metric{Name: "swap.total_bytes", Value: float64(v)})
	}
	if v, ok := info["SwapFree"]; ok {
		out = append(out, metric{Name: "swap.free_bytes", Value: float64(v)})
	}
	if total, ok := info["SwapTotal"]; ok {
		if free, ok2 := info["SwapFree"]; ok2 {
			out = append(out, metric{Name: "swap.used_bytes", Value: float64(total - free)})
		}
	}
	return out, nil
}
