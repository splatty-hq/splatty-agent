package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func parseMeminfo(r io.Reader) map[string]uint64 {
	out := map[string]uint64{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		colon := strings.IndexByte(line, ':')
		if colon == -1 {
			continue
		}
		key := line[:colon]
		fields := strings.Fields(line[colon+1:])
		if len(fields) == 0 {
			continue
		}
		v, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		if len(fields) > 1 && fields[1] == "kB" {
			v *= 1024
		}
		out[key] = v
	}
	return out
}

type memCollector struct {
	procRoot string
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
	return out, nil
}
