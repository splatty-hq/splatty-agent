package main

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type netStat struct {
	rx        uint64
	tx        uint64
	rxPackets uint64
	txPackets uint64
}

func parseNetDev(r io.Reader) map[string]netStat {
	out := map[string]netStat{}
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue
		}
		line := scanner.Text()
		colon := strings.IndexByte(line, ':')
		if colon == -1 {
			continue
		}
		iface := strings.TrimSpace(line[:colon])
		fields := strings.Fields(line[colon+1:])
		if len(fields) < 10 {
			continue
		}
		rx, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		tx, err := strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			continue
		}
		rxPackets, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		txPackets, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			continue
		}
		out[iface] = netStat{rx: rx, tx: tx, rxPackets: rxPackets, txPackets: txPackets}
	}
	return out
}

func netRates(cur, prev map[string]netStat, dt float64) []metric {
	var out []metric
	for iface, s := range cur {
		p, ok := prev[iface]
		if !ok || s.rx < p.rx || s.tx < p.tx || s.rxPackets < p.rxPackets || s.txPackets < p.txPackets {
			continue
		}
		tags := map[string]string{"iface": iface}
		out = append(out,
			metric{Name: "net.rx_bytes_per_sec", Value: float64(s.rx-p.rx) / dt, Tags: tags},
			metric{Name: "net.tx_bytes_per_sec", Value: float64(s.tx-p.tx) / dt, Tags: tags},
			metric{Name: "net.rx_packets_per_sec", Value: float64(s.rxPackets-p.rxPackets) / dt, Tags: tags},
			metric{Name: "net.tx_packets_per_sec", Value: float64(s.txPackets-p.txPackets) / dt, Tags: tags},
		)
	}
	return out
}
