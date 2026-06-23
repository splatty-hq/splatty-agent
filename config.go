package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultSplattyURL = "https://splatty.k0va1.dev"

type config struct {
	URL         string
	Key         string
	Host        string
	Interval    time.Duration
	ProcRoot    string
	SysRoot     string
	DiskMounts  []string
	HTTPTimeout time.Duration
}

func loadConfig() (config, error) {
	key := strings.TrimSpace(os.Getenv("SPLATTY_DSN"))
	if key == "" {
		return config{}, fmt.Errorf("SPLATTY_DSN is required (the key from your project settings)")
	}

	base := strings.TrimRight(envOr("SPLATTY_URL", defaultSplattyURL), "/")
	u, perr := url.Parse(base)
	if perr != nil || u.Scheme == "" || u.Host == "" {
		return config{}, fmt.Errorf("SPLATTY_URL %q is not a valid URL", base)
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
		URL:         fmt.Sprintf("%s/api/metrics", base),
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
