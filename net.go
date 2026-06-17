package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type netStat struct {
	rx uint64
	tx uint64
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
		if len(fields) < 9 {
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
		out[iface] = netStat{rx: rx, tx: tx}
	}
	return out
}

type netCollector struct {
	procRoot string
	prev     map[string]netStat
	prevTime time.Time
}

func (n *netCollector) collect() ([]metric, error) {
	f, err := os.Open(filepath.Join(n.procRoot, "net", "dev"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cur := parseNetDev(f)
	now := time.Now()
	prev := n.prev
	prevTime := n.prevTime
	n.prev = cur
	n.prevTime = now
	if prev == nil {
		return nil, nil
	}
	dt := now.Sub(prevTime).Seconds()
	if dt <= 0 {
		return nil, nil
	}
	var out []metric
	for iface, s := range cur {
		p, ok := prev[iface]
		if !ok {
			continue
		}
		tags := map[string]string{"iface": iface}
		out = append(out,
			metric{Name: "net.rx_bytes_per_sec", Value: float64(s.rx-p.rx) / dt, Tags: tags},
			metric{Name: "net.tx_bytes_per_sec", Value: float64(s.tx-p.tx) / dt, Tags: tags},
		)
	}
	return out, nil
}
