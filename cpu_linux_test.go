//go:build linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCpuCollectorSkipsFirstSampleThenReturnsPercent(t *testing.T) {
	dir := t.TempDir()
	statPath := filepath.Join(dir, "stat")

	if err := os.WriteFile(statPath, []byte("cpu  100 0 50 800 10 0 5 0 0 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &cpuCollector{procRoot: dir}
	ms, err := c.collect()
	if err != nil {
		t.Fatalf("first collect: %v", err)
	}
	if ms != nil {
		t.Fatalf("first collect should be nil (delta needs two samples), got %v", ms)
	}

	if err := os.WriteFile(statPath, []byte("cpu  200 0 100 900 20 0 10 0 0 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ms, err = c.collect()
	if err != nil {
		t.Fatalf("second collect: %v", err)
	}
	if len(ms) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(ms))
	}
	if ms[0].Name != "cpu.usage_percent" {
		t.Errorf("metric name=%q want cpu.usage_percent", ms[0].Name)
	}
	// total delta = 1230-965 = 265, idle delta = (900+20)-(800+10) = 110, busy = 155, pct ≈ 58.49
	want := 58.49
	if got := ms[0].Value; got < want-0.05 || got > want+0.05 {
		t.Errorf("pct=%.2f want=%.2f (±0.05)", got, want)
	}
}
