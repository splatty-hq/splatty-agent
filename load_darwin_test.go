//go:build darwin

package main

import "testing"

func TestLoadDarwinCollectsThreeAverages(t *testing.T) {
	ms, err := newLoadCollector(config{}).collect()
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(ms) != 3 {
		t.Fatalf("expected 3 metrics, got %d", len(ms))
	}
	names := map[string]bool{}
	for _, m := range ms {
		names[m.Name] = true
		if m.Value < 0 {
			t.Errorf("%s value negative: %v", m.Name, m.Value)
		}
	}
	for _, want := range []string{"load.1", "load.5", "load.15"} {
		if !names[want] {
			t.Errorf("missing %q", want)
		}
	}
}

func TestMemDarwinReturnsTotal(t *testing.T) {
	ms, err := newMemCollector(config{}).collect()
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(ms) == 0 {
		t.Fatal("expected at least one metric")
	}
	var sawTotal bool
	for _, m := range ms {
		if m.Name == "mem.total_bytes" {
			sawTotal = true
			if m.Value < 1_000_000_000 {
				t.Errorf("mem.total_bytes=%v looks too small", m.Value)
			}
		}
	}
	if !sawTotal {
		t.Error("missing mem.total_bytes")
	}
}
