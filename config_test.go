package main

import (
	"testing"
)

func TestLoadConfigRequiresMandatoryVars(t *testing.T) {
	t.Setenv("SPLATTY_URL", "")
	t.Setenv("PROJECT_ID", "")
	t.Setenv("PROJECT_KEY", "")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigBuildsURLAndDefaults(t *testing.T) {
	t.Setenv("SPLATTY_URL", "https://splatty.example.com/")
	t.Setenv("PROJECT_ID", "42")
	t.Setenv("PROJECT_KEY", "k")
	t.Setenv("SPLATTY_HOST", "node-1")
	t.Setenv("DISK_MOUNTS", "/, /data")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.URL != "https://splatty.example.com/api/42/metrics" {
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
}

func TestLoadConfigBadInterval(t *testing.T) {
	t.Setenv("SPLATTY_URL", "https://x")
	t.Setenv("PROJECT_ID", "1")
	t.Setenv("PROJECT_KEY", "k")
	t.Setenv("INTERVAL_SECS", "nope")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected error")
	}
}
