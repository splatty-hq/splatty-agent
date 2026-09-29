package main

import (
	"os"
	"testing"
)

func TestParseNetDevSkipsHeadersAndExtractsRxTx(t *testing.T) {
	f, err := os.Open("testdata/proc_net_dev.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got := parseNetDev(f)
	if eth := got["eth0"]; eth.rx != 9876543 || eth.tx != 5555555 {
		t.Errorf("eth0=%+v", eth)
	}
	if eth := got["eth0"]; eth.rxPackets != 54321 || eth.txPackets != 33333 {
		t.Errorf("eth0 packets=%+v", eth)
	}
	if lo := got["lo"]; lo.rx != 1234567 || lo.tx != 1234567 {
		t.Errorf("lo=%+v", lo)
	}
}
