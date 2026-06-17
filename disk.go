//go:build linux || darwin

package main

import "syscall"

type diskCollector struct {
	mounts []string
}

func (d *diskCollector) collect() ([]metric, error) {
	var out []metric
	for _, mount := range d.mounts {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount, &stat); err != nil {
			continue
		}
		blockSize := uint64(stat.Bsize)
		total := uint64(stat.Blocks) * blockSize
		free := uint64(stat.Bavail) * blockSize
		used := total - free
		tags := map[string]string{"mount": mount}
		out = append(out,
			metric{Name: "disk.total_bytes", Value: float64(total), Tags: tags},
			metric{Name: "disk.used_bytes", Value: float64(used), Tags: tags},
			metric{Name: "disk.free_bytes", Value: float64(free), Tags: tags},
		)
	}
	return out, nil
}
