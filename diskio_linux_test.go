//go:build linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiskIOCollectorReportsPhysicalDisksOnly(t *testing.T) {
	procDir := t.TempDir()
	sysDir := t.TempDir()
	blockDir := filepath.Join(sysDir, "block")
	if err := os.MkdirAll(blockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	links := map[string]string{
		"nvme0n1": "../devices/pci0000:00/0000:00:1d.0/nvme/nvme0/nvme0n1",
		"loop0":   "../devices/virtual/block/loop0",
		"dm-0":    "../devices/virtual/block/dm-0",
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(blockDir, name)); err != nil {
			t.Fatal(err)
		}
	}

	statsPath := filepath.Join(procDir, "diskstats")
	write := func(reads, readSectors, writes, writeSectors int) {
		line := fmt.Sprintf(" 259 0 nvme0n1 %d 0 %d 0 %d 0 %d 0 0 0 0\n", reads, readSectors, writes, writeSectors) +
			" 259 1 nvme0n1p1 1 0 1 0 1 0 1 0 0 0 0\n" +
			"   7 0 loop0 1 0 1 0 1 0 1 0 0 0 0\n" +
			" 253 0 dm-0 1 0 1 0 1 0 1 0 0 0 0\n"
		if err := os.WriteFile(statsPath, []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	c := &diskIOCollector{procRoot: procDir, sysRoot: sysDir}
	write(100, 1000, 200, 2000)
	ms, err := c.collect()
	if err != nil || ms != nil {
		t.Fatalf("first collect should be empty, got %v err=%v", ms, err)
	}

	write(110, 1200, 230, 2600)
	c.prevTime = time.Now().Add(-2 * time.Second)
	ms, err = c.collect()
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 4 {
		t.Fatalf("expected 4 metrics for nvme0n1 only, got %d: %v", len(ms), ms)
	}
	want := map[string]float64{
		"disk.read_bytes_per_sec":  200 * 512 / 2.0,
		"disk.write_bytes_per_sec": 600 * 512 / 2.0,
		"disk.read_ops_per_sec":    10 / 2.0,
		"disk.write_ops_per_sec":   30 / 2.0,
	}
	for _, m := range ms {
		if m.Tags["device"] != "nvme0n1" {
			t.Errorf("unexpected device %q", m.Tags["device"])
		}
		w := want[m.Name]
		if m.Value < w*0.95 || m.Value > w*1.05 {
			t.Errorf("%s=%.2f want≈%.2f", m.Name, m.Value, w)
		}
	}
}
