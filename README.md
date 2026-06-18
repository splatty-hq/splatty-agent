# splatty-agent

Single static Go binary. Reads host metrics and POSTs them to a Splatty metrics intake.
Builds for **Linux** (full coverage via `/proc`) and **macOS** (load, mem total, swap, disk
via `sysctl` + `statfs`). Pure Go, no cgo. The only non-stdlib dep is
[`golang.org/x/sys/unix`](https://pkg.go.dev/golang.org/x/sys/unix) for raw sysctl on Darwin.

## Env

| Var | Required | Default | Notes |
|-----|----------|---------|-------|
| `SPLATTY_URL` | yes | — | e.g. `https://splatty.example.com` |
| `PROJECT_ID` | yes | — | numeric project id |
| `PROJECT_KEY` | yes | — | project `ingest_key` |
| `SPLATTY_HOST` | no | hostname | tag value for `host` |
| `INTERVAL_SECS` | no | `15` | collection interval |
| `PROCFS_ROOT` | no | `/proc` | set to `/host/proc` in a container |
| `SYSFS_ROOT` | no | `/sys` | set to `/host/sys` in a container |
| `DISK_MOUNTS` | no | `/` | comma-separated mount paths |

## Run (Docker)

```bash
docker run -d --name splatty-agent --restart unless-stopped \
  -v /proc:/host/proc:ro -v /sys:/host/sys:ro \
  -e PROCFS_ROOT=/host/proc -e SYSFS_ROOT=/host/sys \
  -e SPLATTY_URL=https://splatty.example.com \
  -e PROJECT_ID=<id> -e PROJECT_KEY=<key> \
  ghcr.io/k0va1/splatty-agent:latest
```

Host `/proc` and `/sys` must be mounted into the container; otherwise the agent measures the
container, not the host.

## Run (systemd)

Copy the binary to `/usr/local/bin/splatty-agent` and `systemd/splatty-agent.service` to
`/etc/systemd/system/`. Create `/etc/splatty-agent.env`:

```
SPLATTY_URL=https://splatty.example.com
PROJECT_ID=1
PROJECT_KEY=xxxxxxxxxxxx
```

Then `systemctl enable --now splatty-agent`.

## Metrics emitted

| Metric | Tags | Linux | macOS |
|---|---|:-:|:-:|
| `cpu.usage_percent` | — | ✅ `/proc/stat` | ✅ `top -l 2 -n 0 -s 1` |
| `mem.total_bytes` | — | ✅ `/proc/meminfo` | ✅ `hw.memsize` |
| `mem.used_bytes` | — | ✅ `/proc/meminfo` | ✅ `vm_stat` (active + wired + compressor) |
| `mem.available_bytes` | — | ✅ `/proc/meminfo` | ✅ `vm_stat` (free + inactive + speculative) |
| `swap.total_bytes`, `swap.free_bytes` | — | ✅ `/proc/meminfo` | ✅ `vm.swapusage` |
| `swap.used_bytes` | — | — | ✅ `vm.swapusage` |
| `load.1`, `load.5`, `load.15` | — | ✅ `/proc/loadavg` | ✅ `vm.loadavg` |
| `disk.total_bytes`, `disk.used_bytes`, `disk.free_bytes` | `mount` | ✅ `statfs` | ✅ `statfs` |
| `net.rx_bytes_per_sec`, `net.tx_bytes_per_sec` | `iface` | ✅ `/proc/net/dev` | — |

On macOS, CPU + used/available memory aren't reachable through sysctl; the only sources are
the Mach `host_statistics` API (cgo-only) or the bundled tools. The agent shells out to
`top` (CPU sampling blocks for ~1s) and `vm_stat` instead, both of which ship with every
macOS install. Per-interface network counters on macOS would need
`net.link.generic.ifmibdata` row-walking — not done yet.
