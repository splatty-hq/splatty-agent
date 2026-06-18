//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type cpuCollector struct{}

func newCPUCollector(_ config) collector { return &cpuCollector{} }

var cpuIdleRe = regexp.MustCompile(`([\d.]+)%\s+idle`)

// collect runs `top -l 2 -n 0 -s 1` and reads the second sample's "CPU usage"
// line. The first sample's percentages are bogus (no prior tick reading), so we
// always take the last one. Blocks for ~1s while top samples.
func (c *cpuCollector) collect() ([]metric, error) {
	out, err := exec.Command("/usr/bin/top", "-l", "2", "-n", "0", "-s", "1").Output()
	if err != nil {
		return nil, fmt.Errorf("top: %w", err)
	}
	var lastLine string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "CPU usage:") {
			lastLine = line
		}
	}
	if lastLine == "" {
		return nil, fmt.Errorf("no CPU usage line in top output")
	}
	m := cpuIdleRe.FindStringSubmatch(lastLine)
	if len(m) < 2 {
		return nil, fmt.Errorf("can't parse idle percentage from %q", lastLine)
	}
	idle, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return nil, fmt.Errorf("parse idle: %w", err)
	}
	return []metric{{Name: "cpu.usage_percent", Value: 100.0 - idle}}, nil
}
