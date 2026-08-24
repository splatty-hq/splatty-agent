package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

// version is the release tag, injected at build time with
// -ldflags "-X main.version=<tag>". Empty in builds that don't set it.
var version = ""

// buildVersion reports the version to display: the injected release tag when
// present, otherwise the module version the toolchain records — the tag for
// `go install` builds, the commit it was built from for local ones.
func buildVersion() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
		return bi.Main.Version
	}
	return "unknown"
}

func main() {
	if versionRequested(os.Args[1:]) {
		fmt.Println("splatty-agent", buildVersion())
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("splatty-agent %s starting host=%s interval=%v url=%s", buildVersion(), cfg.Host, cfg.Interval, cfg.URL)

	collectors := []collector{
		newCPUCollector(cfg),
		newMemCollector(cfg),
		newLoadCollector(cfg),
		newDiskCollector(cfg),
		newNetCollector(cfg),
	}
	if cfg.DockerEnabled {
		collectors = append(collectors, newDockerCollector(cfg))
	}

	t := &transport{
		url: cfg.URL,
		key: cfg.Key,
		client: &http.Client{
			Timeout: cfg.HTTPTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        4,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	seenErrors := map[string]bool{}

	for {
		runCycle(cfg, collectors, t, seenErrors)
		select {
		case <-ticker.C:
		case <-sigCh:
			log.Printf("splatty-agent stopping")
			return
		}
	}
}

func versionRequested(args []string) bool {
	for _, a := range args {
		switch a {
		case "--version", "-version", "-v":
			return true
		}
	}
	return false
}

func runCycle(cfg config, collectors []collector, t *transport, seenErrors map[string]bool) {
	b := batch{Host: cfg.Host, Timestamp: time.Now().UTC()}
	for _, c := range collectors {
		ms, err := c.collect()
		if err != nil {
			msg := err.Error()
			if !seenErrors[msg] {
				log.Printf("collect error (logging once): %v", err)
				seenErrors[msg] = true
			}
			continue
		}
		b.Metrics = append(b.Metrics, ms...)
	}
	if len(b.Metrics) == 0 {
		return
	}
	if err := t.send(b); err != nil {
		log.Printf("send error: %v", err)
	}
}
