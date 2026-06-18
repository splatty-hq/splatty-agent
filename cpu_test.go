package main

import (
	"os"
	"strings"
	"testing"
)

func TestParseProcStatSumsAllFieldsAndExtractsIdleIowait(t *testing.T) {
	f, err := os.Open("testdata/proc_stat.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got, err := parseProcStat(f)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	wantTotal := uint64(100 + 0 + 50 + 800 + 10 + 0 + 5 + 0 + 0 + 0)
	wantIdle := uint64(800 + 10)
	if got.total != wantTotal {
		t.Errorf("total=%d want=%d", got.total, wantTotal)
	}
	if got.idle != wantIdle {
		t.Errorf("idle=%d want=%d", got.idle, wantIdle)
	}
}

func TestParseProcStatMissingCpuLine(t *testing.T) {
	_, err := parseProcStat(strings.NewReader("intr 5\nctxt 9\n"))
	if err == nil {
		t.Fatal("expected error on missing cpu line")
	}
}
