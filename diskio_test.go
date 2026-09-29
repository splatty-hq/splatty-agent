package main

import (
	"os"
	"testing"
)

func TestParseDiskstatsExtractsOpsAndSectors(t *testing.T) {
	f, err := os.Open("testdata/proc_diskstats.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got := parseDiskstats(f)
	want := diskIOStat{reads: 1000, readSectors: 80000, writes: 2000, writeSectors: 160000}
	if nvme := got["nvme0n1"]; nvme != want {
		t.Errorf("nvme0n1=%+v want %+v", nvme, want)
	}
	if len(got) != 4 {
		t.Errorf("expected 4 devices, got %d", len(got))
	}
}
