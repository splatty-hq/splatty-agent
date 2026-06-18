package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseLoadavg(s string) (load1, load5, load15 float64, err error) {
	fields := strings.Fields(s)
	if len(fields) < 3 {
		err = fmt.Errorf("loadavg too short")
		return
	}
	if load1, err = strconv.ParseFloat(fields[0], 64); err != nil {
		return
	}
	if load5, err = strconv.ParseFloat(fields[1], 64); err != nil {
		return
	}
	if load15, err = strconv.ParseFloat(fields[2], 64); err != nil {
		return
	}
	return
}
