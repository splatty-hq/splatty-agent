//go:build darwin

package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

type netCollector struct {
	prev     map[string]netStat
	prevTime time.Time
}

func newNetCollector(_ config) collector { return &netCollector{} }

func (n *netCollector) collect() ([]metric, error) {
	buf, err := routeIfList2()
	if err != nil {
		return nil, fmt.Errorf("sysctl NET_RT_IFLIST2: %w", err)
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}
	names := make(map[uint16]string, len(ifaces))
	for _, ifc := range ifaces {
		names[uint16(ifc.Index)] = ifc.Name
	}

	cur := map[string]netStat{}
	for idx, s := range parseIfList2(buf) {
		if name := names[idx]; isPhysicalDarwinIface(name) {
			cur[name] = s
		}
	}
	now := time.Now()
	prev := n.prev
	prevTime := n.prevTime
	n.prev = cur
	n.prevTime = now
	if prev == nil {
		return nil, nil
	}
	dt := now.Sub(prevTime).Seconds()
	if dt <= 0 {
		return nil, nil
	}
	return netRates(cur, prev, dt), nil
}

// routeIfList2 fetches the interface list from the routing sysctl. The
// `net.route` node has no name x/sys/unix can resolve, and golang.org/x/net/route
// (which syscall.RouteRIB's deprecation points at) doesn't expose the if_data64
// counters, so this is the one cgo-free way in.
func routeIfList2() ([]byte, error) {
	return syscall.RouteRIB(unix.NET_RT_IFLIST2, 0)
}

// isPhysicalDarwinIface keeps the en* Ethernet, Wi-Fi and Thunderbolt ports.
// Everything else on macOS is virtual (lo0, utun VPN tunnels, awdl/llw AirDrop,
// bridge0 over the Thunderbolt ports, anpi) and would count the same traffic
// twice.
func isPhysicalDarwinIface(name string) bool {
	return strings.HasPrefix(name, "en")
}

// parseIfList2 walks the RTM_IFINFO2 messages a NET_RT_IFLIST2 sysctl returns,
// keyed by interface index. Their if_data64 counters are 64-bit, unlike the
// 32-bit if_data in NET_RT_IFLIST that wraps at 4 GiB.
func parseIfList2(buf []byte) map[uint16]netStat {
	out := map[uint16]netStat{}
	for len(buf) >= 4 {
		msglen := int(binary.LittleEndian.Uint16(buf))
		if msglen < 4 || msglen > len(buf) {
			break
		}
		if buf[3] == unix.RTM_IFINFO2 && msglen >= unix.SizeofIfMsghdr2 {
			var m unix.IfMsghdr2
			copy((*[unix.SizeofIfMsghdr2]byte)(unsafe.Pointer(&m))[:], buf)
			out[m.Index] = netStat{
				rx:        m.Data.Ibytes,
				tx:        m.Data.Obytes,
				rxPackets: m.Data.Ipackets,
				txPackets: m.Data.Opackets,
			}
		}
		buf = buf[msglen:]
	}
	return out
}
