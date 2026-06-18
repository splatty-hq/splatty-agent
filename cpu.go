package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type cpuStat struct {
	total uint64
	idle  uint64
}

func parseProcStat(r io.Reader) (cpuStat, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		var total, idle uint64
		for i, f := range fields[1:] {
			v, err := strconv.ParseUint(f, 10, 64)
			if err != nil {
				return cpuStat{}, fmt.Errorf("parse cpu field %d: %w", i, err)
			}
			total += v
			if i == 3 || i == 4 {
				idle += v
			}
		}
		return cpuStat{total: total, idle: idle}, nil
	}
	if err := scanner.Err(); err != nil {
		return cpuStat{}, err
	}
	return cpuStat{}, fmt.Errorf("cpu line not found")
}
