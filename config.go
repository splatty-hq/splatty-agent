package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultSplattyURL = "https://splatty.app"
const defaultDockerSocket = "/var/run/docker.sock"

type config struct {
	URL           string
	Key           string
	Host          string
	Interval      time.Duration
	ProcRoot      string
	SysRoot       string
	DiskMounts    []string
	HTTPTimeout   time.Duration
	DockerEnabled bool
	DockerSocket  string
}

func loadConfig() (config, error) {
	key := strings.TrimSpace(os.Getenv("SPLATTY_SERVER_TOKEN"))
	if key == "" {
		return config{}, fmt.Errorf("SPLATTY_SERVER_TOKEN is required (the agent token from the server's settings page)")
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
		URL:           fmt.Sprintf("%s/api/metrics", base),
		Key:           key,
		Host:          host,
		Interval:      time.Duration(intervalSecs) * time.Second,
		ProcRoot:      envOr("PROCFS_ROOT", "/proc"),
		SysRoot:       envOr("SYSFS_ROOT", "/sys"),
		DiskMounts:    mounts,
		HTTPTimeout:   10 * time.Second,
		DockerEnabled: envBool("DOCKER_ENABLED", false),
		DockerSocket:  envOr("DOCKER_SOCKET", defaultDockerSocket),
	}, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envBool(k string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
	switch v {
	case "":
		return def
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
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
