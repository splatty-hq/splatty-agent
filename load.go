package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func parseLoadavg(s string) (load1, load5, load15 float64, err error) {
	fields := strings.Fields(s)
	if len(fields) < 3 {
		err = fmt.Errorf("loadavg too short")
		return
	}
	if load1, err = strconv.ParseFloat(fields[0], 64); err != nil {
		return
	}
	if load5, err = strconv.ParseFloat(fields[1], 64); err != nil {
		return
	}
	if load15, err = strconv.ParseFloat(fields[2], 64); err != nil {
		return
	}
	return
}

type loadCollector struct {
	procRoot string
}

func (l *loadCollector) collect() ([]metric, error) {
	b, err := os.ReadFile(filepath.Join(l.procRoot, "loadavg"))
	if err != nil {
		return nil, err
	}
	one, five, fifteen, err := parseLoadavg(string(b))
	if err != nil {
		return nil, err
	}
	return []metric{
		{Name: "load.1", Value: one},
		{Name: "load.5", Value: five},
		{Name: "load.15", Value: fifteen},
	}, nil
}
