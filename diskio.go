package main

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

const diskSectorBytes = 512

type diskIOStat struct {
	reads        uint64
	readSectors  uint64
	writes       uint64
	writeSectors uint64
}

// parseDiskstats reads /proc/diskstats. Sector counts there are always in
// 512-byte units, regardless of the device's physical sector size.
func parseDiskstats(r io.Reader) map[string]diskIOStat {
	out := map[string]diskIOStat{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		var vals [4]uint64
		ok := true
		for i, idx := range []int{3, 5, 7, 9} {
			v, err := strconv.ParseUint(fields[idx], 10, 64)
			if err != nil {
				ok = false
				break
			}
			vals[i] = v
		}
		if !ok {
			continue
		}
		out[fields[2]] = diskIOStat{reads: vals[0], readSectors: vals[1], writes: vals[2], writeSectors: vals[3]}
	}
	return out
}
