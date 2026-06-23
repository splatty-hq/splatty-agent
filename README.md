# splatty-agent

Single static Go binary. Reads host metrics and POSTs them to a Splatty metrics intake.
Builds for **Linux** (full coverage via `/proc`) and **macOS** (load, mem total, swap, disk
via `sysctl` + `statfs`). Pure Go, no cgo. The only non-stdlib dep is
[`golang.org/x/sys/unix`](https://pkg.go.dev/golang.org/x/sys/unix) for raw sysctl on Darwin.

## Env

| Var | Required | Default | Notes |
|-----|----------|---------|-------|
| `SPLATTY_DSN` | yes | — | hex key from the project settings page |
| `SPLATTY_URL` | no | `https://splatty.k0va1.dev` | server URL |
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
  -e SPLATTY_DSN=<hex> \
  -e SPLATTY_URL=https://splatty.k0va1.dev \
  ghcr.io/splatty-hq/splatty-agent:latest
```

Host `/proc` and `/sys` must be mounted into the container; otherwise the agent measures the
container, not the host.

## Install (prebuilt binary)

Released for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`. The install
script picks the right asset, verifies its `.sha256`, and drops the binary at
`/usr/local/bin/splatty-agent`.

Binary only:

```bash
curl -fsSL https://raw.githubusercontent.com/splatty-hq/splatty-agent/master/install.sh | sh
```

Linux + systemd, all-in-one (binary + service that survives reboot):

```bash
curl -fsSL https://raw.githubusercontent.com/splatty-hq/splatty-agent/master/install.sh \
  | sudo SPLATTY_DSN=<hex> sh
```

When `SPLATTY_DSN` is set on a Linux host with systemd, the script also creates a
`splatty` system user, writes `/etc/splatty-agent.env` (`0640 root:splatty`), installs
`splatty-agent.service`, and `systemctl enable --now`s it. Without it, only the binary
is installed.

Optional overrides:

```bash
VERSION=v0.1.0           # pin a release tag (default: latest)
INSTALL_DIR=$HOME/.local/bin   # binary destination (default: /usr/local/bin)
SPLATTY_HOST=...         # passed through to the service env file
INTERVAL_SECS=...
DISK_MOUNTS=...
```

Check the service: `systemctl status splatty-agent` · logs: `journalctl -u splatty-agent -f`.

## Run (systemd, manual)

If you'd rather not run the script, the unit file ships in the repo. Copy the binary to
`/usr/local/bin/splatty-agent` and `systemd/splatty-agent.service` to `/etc/systemd/system/`.
Create the `splatty` user (`useradd --system --no-create-home --shell /usr/sbin/nologin splatty`)
and write `/etc/splatty-agent.env`:

```
SPLATTY_DSN=<hex>
SPLATTY_URL=https://splatty.k0va1.dev
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
