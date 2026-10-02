package main

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const diskSectorBytes = 512

type diskIOStat struct {
	reads      uint64
	readBytes  uint64
	writes     uint64
	writeBytes uint64
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
		out[fields[2]] = diskIOStat{
			reads:      vals[0],
			readBytes:  vals[1] * diskSectorBytes,
			writes:     vals[2],
			writeBytes: vals[3] * diskSectorBytes,
		}
	}
	return out
}

var (
	ioregInterconnectRe = regexp.MustCompile(`"Physical Interconnect"="([^"]*)"`)
	ioregBSDNameRe      = regexp.MustCompile(`"BSD Name" = "([^"]+)"`)
	ioregStatRe         = regexp.MustCompile(`"(Bytes|Operations) \((Read|Write)\)"=(\d+)`)
)

// parseIORegBlockStorage reads `ioreg -r -c IOBlockStorageDevice -d 3 -l -w 0`.
// Each top-level entry is a storage device, with its driver's "Statistics" and
// the whole-disk IOMedia's "BSD Name" nested beneath it. Disk images report a
// "Virtual Interface" interconnect and are skipped, as are devices with no
// media inserted.
func parseIORegBlockStorage(r io.Reader) map[string]diskIOStat {
	out := map[string]diskIOStat{}
	var (
		name    string
		virtual bool
		stat    diskIOStat
		hasStat bool
	)
	flush := func() {
		if name != "" && hasStat && !virtual {
			out[name] = stat
		}
		name, virtual, stat, hasStat = "", false, diskIOStat{}, false
	}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+-o ") {
			flush()
			continue
		}
		if m := ioregInterconnectRe.FindStringSubmatch(line); m != nil {
			virtual = m[1] == "Virtual Interface"
		}
		if m := ioregBSDNameRe.FindStringSubmatch(line); m != nil && name == "" {
			name = m[1]
		}
		if !strings.Contains(line, `"Statistics" = {`) || hasStat {
			continue
		}
		hasStat = true
		for _, m := range ioregStatRe.FindAllStringSubmatch(line, -1) {
			v, err := strconv.ParseUint(m[3], 10, 64)
			if err != nil {
				continue
			}
			switch m[1] + " " + m[2] {
			case "Bytes Read":
				stat.readBytes = v
			case "Bytes Write":
				stat.writeBytes = v
			case "Operations Read":
				stat.reads = v
			case "Operations Write":
				stat.writes = v
			}
		}
	}
	flush()
	return out
}

func diskIORates(cur, prev map[string]diskIOStat, dt float64) []metric {
	var out []metric
	for dev, s := range cur {
		p, ok := prev[dev]
		if !ok || s.reads < p.reads || s.writes < p.writes || s.readBytes < p.readBytes || s.writeBytes < p.writeBytes {
			continue
		}
		tags := map[string]string{"device": dev}
		out = append(out,
			metric{Name: "disk.read_bytes_per_sec", Value: float64(s.readBytes-p.readBytes) / dt, Tags: tags},
			metric{Name: "disk.write_bytes_per_sec", Value: float64(s.writeBytes-p.writeBytes) / dt, Tags: tags},
			metric{Name: "disk.read_ops_per_sec", Value: float64(s.reads-p.reads) / dt, Tags: tags},
			metric{Name: "disk.write_ops_per_sec", Value: float64(s.writes-p.writes) / dt, Tags: tags},
		)
	}
	return out
}
