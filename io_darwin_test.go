//go:build darwin

package main

import (
	"testing"
	"time"
)

func TestParseIfList2FindsLoopback(t *testing.T) {
	buf, err := routeIfList2()
	if err != nil {
		t.Fatal(err)
	}
	got := parseIfList2(buf)
	if len(got) == 0 {
		t.Fatal("no interfaces parsed")
	}
	if lo, ok := got[1]; !ok || lo.rxPackets == 0 {
		t.Errorf("lo0 (index 1)=%+v ok=%v", lo, ok)
	}
}

func TestNetDarwinReportsOnlyPhysicalIfaces(t *testing.T) {
	c := &netCollector{}
	if ms, err := c.collect(); err != nil || ms != nil {
		t.Fatalf("first collect should be empty, got %v err=%v", ms, err)
	}
	c.prevTime = time.Now().Add(-time.Second)
	ms, err := c.collect()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range ms {
		if !isPhysicalDarwinIface(m.Tags["iface"]) {
			t.Errorf("virtual iface %q reported", m.Tags["iface"])
		}
		if m.Value < 0 {
			t.Errorf("%s negative: %v", m.Name, m.Value)
		}
	}
}

func TestDiskIODarwinReportsWholeDisks(t *testing.T) {
	c := &diskIOCollector{}
	if ms, err := c.collect(); err != nil || ms != nil {
		t.Fatalf("first collect should be empty, got %v err=%v", ms, err)
	}
	c.prevTime = time.Now().Add(-time.Second)
	ms, err := c.collect()
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) == 0 {
		t.Fatal("expected disk I/O metrics for at least one disk")
	}
	for _, m := range ms {
		if m.Tags["device"] == "" || m.Value < 0 {
			t.Errorf("bad metric %+v", m)
		}
	}
}
