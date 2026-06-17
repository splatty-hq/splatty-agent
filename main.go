package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("splatty-agent starting host=%s interval=%v url=%s", cfg.Host, cfg.Interval, cfg.URL)

	collectors := []collector{
		&cpuCollector{procRoot: cfg.ProcRoot},
		&memCollector{procRoot: cfg.ProcRoot},
		&loadCollector{procRoot: cfg.ProcRoot},
		&diskCollector{mounts: cfg.DiskMounts},
		&netCollector{procRoot: cfg.ProcRoot},
	}

	t := &transport{
		url:    cfg.URL,
		key:    cfg.Key,
		client: &http.Client{Timeout: cfg.HTTPTimeout},
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		runCycle(cfg, collectors, t)
		select {
		case <-ticker.C:
		case <-sigCh:
			log.Printf("splatty-agent stopping")
			return
		}
	}
}

func runCycle(cfg config, collectors []collector, t *transport) {
	b := batch{Host: cfg.Host, Timestamp: time.Now().UTC()}
	for _, c := range collectors {
		ms, err := c.collect()
		if err != nil {
			log.Printf("collect error: %v", err)
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
