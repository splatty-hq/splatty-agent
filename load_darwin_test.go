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

func TestMemDarwinReturnsTotalUsedAvailable(t *testing.T) {
	ms, err := newMemCollector(config{}).collect()
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	want := map[string]bool{"mem.total_bytes": false, "mem.used_bytes": false, "mem.available_bytes": false}
	for _, m := range ms {
		if _, ok := want[m.Name]; ok {
			want[m.Name] = true
			if m.Value < 1_000_000 {
				t.Errorf("%s=%v looks too small", m.Name, m.Value)
			}
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("missing %s", k)
		}
	}
}

func TestParseVMStat(t *testing.T) {
	input := `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                               1052.
Pages active:                             47931.
Pages inactive:                           3122.
Pages speculative:                        66.
Pages wired down:                         52384.
Pages occupied by compressor:             92571.
"Translation faults":                     12345.
`
	got := parseVMStat(input)
	if got["Pages free"] != 1052 {
		t.Errorf("Pages free=%d", got["Pages free"])
	}
	if got["Pages active"] != 47931 {
		t.Errorf("Pages active=%d", got["Pages active"])
	}
	if got["Pages occupied by compressor"] != 92571 {
		t.Errorf("compressor=%d", got["Pages occupied by compressor"])
	}
}

func TestCpuDarwinReturnsUsagePercent(t *testing.T) {
	ms, err := newCPUCollector(config{}).collect()
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(ms) != 1 || ms[0].Name != "cpu.usage_percent" {
		t.Fatalf("expected cpu.usage_percent, got %+v", ms)
	}
	if v := ms[0].Value; v < 0 || v > 100 {
		t.Errorf("usage=%v out of [0,100]", v)
	}
}
