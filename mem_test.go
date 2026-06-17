package main

import (
	"os"
	"testing"
)

func TestParseMeminfoConvertsKbToBytes(t *testing.T) {
	f, err := os.Open("testdata/proc_meminfo.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got := parseMeminfo(f)
	if got["MemTotal"] != 16384000*1024 {
		t.Errorf("MemTotal=%d", got["MemTotal"])
	}
	if got["MemAvailable"] != 8192000*1024 {
		t.Errorf("MemAvailable=%d", got["MemAvailable"])
	}
	if got["SwapFree"] != 2097152*1024 {
		t.Errorf("SwapFree=%d", got["SwapFree"])
	}
}
