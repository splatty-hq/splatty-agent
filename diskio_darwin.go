//go:build darwin

package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// IOKit's per-disk counters need cgo to read directly, so the collector parses
// the same IOBlockStorageDriver "Statistics" through the bundled ioreg tool.
type diskIOCollector struct {
	prev     map[string]diskIOStat
	prevTime time.Time
}

func newDiskIOCollector(_ config) collector { return &diskIOCollector{} }

func (d *diskIOCollector) collect() ([]metric, error) {
	raw, err := exec.Command("/usr/sbin/ioreg", "-r", "-c", "IOBlockStorageDevice", "-d", "3", "-l", "-w", "0").Output()
	if err != nil {
		return nil, fmt.Errorf("ioreg: %w", err)
	}
	cur := parseIORegBlockStorage(bytes.NewReader(raw))
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
	return diskIORates(cur, prev, dt), nil
}
