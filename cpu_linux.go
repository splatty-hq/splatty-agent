//go:build linux

package main

import (
	"os"
	"path/filepath"
)

type cpuCollector struct {
	procRoot string
	prev     *cpuStat
}

func newCPUCollector(cfg config) collector {
	return &cpuCollector{procRoot: cfg.ProcRoot}
}

func (c *cpuCollector) collect() ([]metric, error) {
	f, err := os.Open(filepath.Join(c.procRoot, "stat"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cur, err := parseProcStat(f)
	if err != nil {
		return nil, err
	}
	prev := c.prev
	c.prev = &cur
	if prev == nil {
		return nil, nil
	}

	totalDelta := cur.total - prev.total
	idleDelta := cur.idle - prev.idle
	if totalDelta == 0 {
		return nil, nil
	}
	pct := float64(totalDelta-idleDelta) / float64(totalDelta) * 100.0
	return []metric{{Name: "cpu.usage_percent", Value: pct}}, nil
}
