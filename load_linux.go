//go:build linux

package main

import (
	"os"
	"path/filepath"
)

type loadCollector struct {
	procRoot string
}

func newLoadCollector(cfg config) collector {
	return &loadCollector{procRoot: cfg.ProcRoot}
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
