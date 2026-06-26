package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const projectLabel = "com.splatty.project"

type dockerCollector struct {
	socket string
	client *http.Client
}

func newDockerCollector(cfg config) collector {
	tr := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", cfg.DockerSocket)
		},
	}
	return &dockerCollector{
		socket: cfg.DockerSocket,
		client: &http.Client{Timeout: 5 * time.Second, Transport: tr},
	}
}

type dockerContainer struct {
	ID     string            `json:"Id"`
	Names  []string          `json:"Names"`
	Image  string            `json:"Image"`
	Labels map[string]string `json:"Labels"`
}

type dockerStats struct {
	CPUStats    cpuStats    `json:"cpu_stats"`
	PreCPUStats cpuStats    `json:"precpu_stats"`
	MemoryStats memoryStats `json:"memory_stats"`
	Networks    map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
	BlkioStats struct {
		IOServiceBytesRecursive []struct {
			Op    string `json:"op"`
			Value uint64 `json:"value"`
		} `json:"io_service_bytes_recursive"`
	} `json:"blkio_stats"`
}

type cpuStats struct {
	CPUUsage struct {
		TotalUsage uint64 `json:"total_usage"`
	} `json:"cpu_usage"`
	SystemUsage uint64 `json:"system_cpu_usage"`
	OnlineCPUs  uint64 `json:"online_cpus"`
}

type memoryStats struct {
	Usage uint64 `json:"usage"`
	Limit uint64 `json:"limit"`
}

func (d *dockerCollector) collect() ([]metric, error) {
	containers, err := d.listContainers()
	if err != nil {
		return nil, fmt.Errorf("docker list: %w", err)
	}
	var out []metric
	for _, c := range containers {
		slug := strings.TrimSpace(c.Labels[projectLabel])
		if slug == "" {
			continue
		}
		stats, err := d.fetchStats(c.ID)
		if err != nil {
			continue
		}
		name := containerName(c)
		tags := map[string]string{
			"container":    name,
			"image":        c.Image,
			"project_slug": slug,
		}
		out = append(out, statsToMetrics(stats, tags)...)
	}
	return out, nil
}

func (d *dockerCollector) listContainers() ([]dockerContainer, error) {
	resp, err := d.client.Get("http://docker/containers/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var cs []dockerContainer
	if err := json.NewDecoder(resp.Body).Decode(&cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func (d *dockerCollector) fetchStats(id string) (dockerStats, error) {
	var s dockerStats
	resp, err := d.client.Get("http://docker/containers/" + id + "/stats?stream=false&one-shot=true")
	if err != nil {
		return s, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return s, fmt.Errorf("status %d", resp.StatusCode)
	}
	return s, json.NewDecoder(resp.Body).Decode(&s)
}

func containerName(c dockerContainer) string {
	if len(c.Names) == 0 {
		return c.ID[:12]
	}
	return strings.TrimPrefix(c.Names[0], "/")
}

func statsToMetrics(s dockerStats, tags map[string]string) []metric {
	out := []metric{
		{Name: "container.mem.used_bytes", Value: float64(s.MemoryStats.Usage), Tags: tags},
		{Name: "container.mem.limit_bytes", Value: float64(s.MemoryStats.Limit), Tags: tags},
	}

	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemUsage) - float64(s.PreCPUStats.SystemUsage)
	if cpuDelta > 0 && sysDelta > 0 {
		cores := float64(s.CPUStats.OnlineCPUs)
		if cores == 0 {
			cores = 1
		}
		pct := cpuDelta / sysDelta * cores * 100.0
		out = append(out, metric{Name: "container.cpu.usage_percent", Value: pct, Tags: tags})
	}

	var rx, tx uint64
	for _, n := range s.Networks {
		rx += n.RxBytes
		tx += n.TxBytes
	}
	out = append(out,
		metric{Name: "container.net.rx_bytes", Value: float64(rx), Tags: tags},
		metric{Name: "container.net.tx_bytes", Value: float64(tx), Tags: tags},
	)

	var read, write uint64
	for _, e := range s.BlkioStats.IOServiceBytesRecursive {
		switch e.Op {
		case "Read", "read":
			read += e.Value
		case "Write", "write":
			write += e.Value
		}
	}
	out = append(out,
		metric{Name: "container.block.read_bytes", Value: float64(read), Tags: tags},
		metric{Name: "container.block.write_bytes", Value: float64(write), Tags: tags},
	)
	return out
}
