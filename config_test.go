package main

import (
	"testing"
)

func TestLoadConfigRequiresToken(t *testing.T) {
	t.Setenv("SPLATTY_SERVER_TOKEN", "")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigDefaultsURL(t *testing.T) {
	t.Setenv("SPLATTY_SERVER_TOKEN", "abc123def456")
	t.Setenv("SPLATTY_URL", "")
	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.URL != "https://splatty.app/api/metrics" {
		t.Errorf("URL=%q", cfg.URL)
	}
	if cfg.Key != "abc123def456" {
		t.Errorf("Key=%q", cfg.Key)
	}
	if cfg.DockerEnabled {
		t.Errorf("DockerEnabled defaults to true, expected false")
	}
}

func TestLoadConfigCustomURL(t *testing.T) {
	t.Setenv("SPLATTY_SERVER_TOKEN", "abc123")
	t.Setenv("SPLATTY_URL", "http://localhost:3000/")
	t.Setenv("SPLATTY_HOST", "node-1")
	t.Setenv("DISK_MOUNTS", "/, /data")
	t.Setenv("DOCKER_ENABLED", "true")
	t.Setenv("DOCKER_SOCKET", "/run/docker.sock")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.URL != "http://localhost:3000/api/metrics" {
		t.Errorf("URL=%q", cfg.URL)
	}
	if cfg.Host != "node-1" {
		t.Errorf("Host=%q", cfg.Host)
	}
	if len(cfg.DiskMounts) != 2 || cfg.DiskMounts[1] != "/data" {
		t.Errorf("Mounts=%v", cfg.DiskMounts)
	}
	if cfg.Interval.Seconds() != 15 {
		t.Errorf("Interval=%v", cfg.Interval)
	}
	if !cfg.DockerEnabled {
		t.Errorf("DockerEnabled=%v, expected true", cfg.DockerEnabled)
	}
	if cfg.DockerSocket != "/run/docker.sock" {
		t.Errorf("DockerSocket=%q", cfg.DockerSocket)
	}
}

func TestLoadConfigBadURL(t *testing.T) {
	t.Setenv("SPLATTY_SERVER_TOKEN", "abc123")
	t.Setenv("SPLATTY_URL", "not-a-url")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigBadInterval(t *testing.T) {
	t.Setenv("SPLATTY_SERVER_TOKEN", "abc123")
	t.Setenv("INTERVAL_SECS", "nope")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected error")
	}
}
