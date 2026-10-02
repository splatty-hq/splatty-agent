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
	want := diskIOStat{reads: 1000, readBytes: 80000 * diskSectorBytes, writes: 2000, writeBytes: 160000 * diskSectorBytes}
	if nvme := got["nvme0n1"]; nvme != want {
		t.Errorf("nvme0n1=%+v want %+v", nvme, want)
	}
	if len(got) != 4 {
		t.Errorf("expected 4 devices, got %d", len(got))
	}
}

func TestParseIORegBlockStorageSkipsDiskImagesAndEmptyReaders(t *testing.T) {
	f, err := os.Open("testdata/ioreg_block_storage.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got := parseIORegBlockStorage(f)
	want := diskIOStat{reads: 364074635, readBytes: 6549502758912, writes: 452881355, writeBytes: 6097090019328}
	if disk0 := got["disk0"]; disk0 != want {
		t.Errorf("disk0=%+v want %+v", disk0, want)
	}
	if len(got) != 1 {
		t.Errorf("expected only disk0, got %+v", got)
	}
}

func TestDiskIORatesSkipsNewAndResetDevices(t *testing.T) {
	prev := map[string]diskIOStat{
		"disk0": {reads: 10, readBytes: 1000, writes: 20, writeBytes: 2000},
		"disk1": {reads: 50, readBytes: 5000, writes: 50, writeBytes: 5000},
	}
	cur := map[string]diskIOStat{
		"disk0": {reads: 14, readBytes: 1800, writes: 26, writeBytes: 3000},
		"disk1": {reads: 1, readBytes: 100, writes: 1, writeBytes: 100},
		"disk2": {reads: 1, readBytes: 100, writes: 1, writeBytes: 100},
	}
	got := map[string]float64{}
	for _, m := range diskIORates(cur, prev, 2) {
		if m.Tags["device"] != "disk0" {
			t.Errorf("unexpected device %q", m.Tags["device"])
		}
		got[m.Name] = m.Value
	}
	want := map[string]float64{
		"disk.read_bytes_per_sec":  400,
		"disk.write_bytes_per_sec": 500,
		"disk.read_ops_per_sec":    2,
		"disk.write_ops_per_sec":   3,
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("%s=%v want %v", name, got[name], w)
		}
	}
}
