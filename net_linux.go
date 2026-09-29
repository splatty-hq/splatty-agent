//go:build linux

package main

import (
	"os"
	"path/filepath"
	"time"
)

type netCollector struct {
	procRoot string
	prev     map[string]netStat
	prevTime time.Time
}

func newNetCollector(cfg config) collector {
	return &netCollector{procRoot: cfg.ProcRoot}
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
	return out, nil
}
