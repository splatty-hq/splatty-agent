package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	URL        string
	ProjectID  string
	Key        string
	Host       string
	Interval   time.Duration
	ProcRoot   string
	SysRoot    string
	DiskMounts []string
	HTTPTimeout time.Duration
}

func loadConfig() (config, error) {
	baseURL := strings.TrimRight(os.Getenv("SPLATTY_URL"), "/")
	projectID := os.Getenv("PROJECT_ID")
	key := os.Getenv("PROJECT_KEY")
	if baseURL == "" || projectID == "" || key == "" {
		return config{}, fmt.Errorf("SPLATTY_URL, PROJECT_ID and PROJECT_KEY are required")
	}

	intervalSecs := 15
	if v := os.Getenv("INTERVAL_SECS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return config{}, fmt.Errorf("INTERVAL_SECS must be a positive integer, got %q", v)
		}
		intervalSecs = n
	}

	host := os.Getenv("SPLATTY_HOST")
	if host == "" {
		host, _ = os.Hostname()
	}

	mounts := splitCSV(envOr("DISK_MOUNTS", "/"))

	return config{
		URL:         fmt.Sprintf("%s/api/%s/metrics", baseURL, projectID),
		ProjectID:   projectID,
		Key:         key,
		Host:        host,
		Interval:    time.Duration(intervalSecs) * time.Second,
		ProcRoot:    envOr("PROCFS_ROOT", "/proc"),
		SysRoot:     envOr("SYSFS_ROOT", "/sys"),
		DiskMounts:  mounts,
		HTTPTimeout: 10 * time.Second,
	}, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
