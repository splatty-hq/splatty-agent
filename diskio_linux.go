//go:build linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type diskIOCollector struct {
	procRoot string
	sysRoot  string
	prev     map[string]diskIOStat
	prevTime time.Time
}

func newDiskIOCollector(cfg config) collector {
	return &diskIOCollector{procRoot: cfg.ProcRoot, sysRoot: cfg.SysRoot}
}

func (d *diskIOCollector) collect() ([]metric, error) {
	f, err := os.Open(filepath.Join(d.procRoot, "diskstats"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cur := parseDiskstats(f)
	now := time.Now()
	prev := d.prev
	prevTime := d.prevTime
	d.prev = cur
	d.prevTime = now
	if prev == nil {
		return nil, nil
	}
	dt := now.Sub(prevTime).Seconds()
	if dt <= 0 {
		return nil, nil
	}
	physical := d.physicalDisks()
	for dev := range cur {
		if !physical[dev] {
			delete(cur, dev)
		}
	}
	return diskIORates(cur, prev, dt), nil
}

// physicalDisks lists whole block devices backed by hardware. Partitions never
// appear in /sys/block, and loop, ram, zram, dm and md devices resolve under
// /sys/devices/virtual, so skipping both keeps stacked devices from counting
// the same I/O twice.
func (d *diskIOCollector) physicalDisks() map[string]bool {
	blockDir := filepath.Join(d.sysRoot, "block")
	entries, err := os.ReadDir(blockDir)
	if err != nil {
		return nil
	}
	out := map[string]bool{}
	for _, e := range entries {
		target, err := os.Readlink(filepath.Join(blockDir, e.Name()))
		if err != nil || strings.Contains(target, "/devices/virtual/") {
			continue
		}
		out[e.Name()] = true
	}
	return out
}
