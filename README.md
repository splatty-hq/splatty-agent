# splatty-agent

Single static Go binary. Reads host metrics and POSTs them to a Splatty metrics intake.
Builds for **Linux** (full coverage via `/proc`) and **macOS** (`sysctl`, `statfs`, and the
bundled `top`/`vm_stat`/`ioreg` tools). Pure Go, no cgo. The only non-stdlib dep is
[`golang.org/x/sys/unix`](https://pkg.go.dev/golang.org/x/sys/unix) for raw sysctl on Darwin.

## Env

| Var | Required | Default | Notes |
|-----|----------|---------|-------|
| `SPLATTY_SERVER_TOKEN` | yes | — | agent token from the server's settings page |
| `SPLATTY_URL` | no | `https://splatty.app` | server URL |
| `SPLATTY_HOST` | no | hostname | tag value for `host` |
| `INTERVAL_SECS` | no | `15` | collection interval |
| `PROCFS_ROOT` | no | `/proc` | set to `/host/proc` in a container |
| `SYSFS_ROOT` | no | `/sys` | set to `/host/sys` in a container |
| `DISK_MOUNTS` | no | `/` | comma-separated mount paths |
| `DOCKER_ENABLED` | no | `false` | set to `true` to collect per-container metrics |
| `DOCKER_SOCKET` | no | `/var/run/docker.sock` | docker engine socket path |

## Run (Docker)

```bash
docker run -d --name splatty-agent --restart unless-stopped \
  -v /proc:/host/proc:ro -v /sys:/host/sys:ro \
  -e PROCFS_ROOT=/host/proc -e SYSFS_ROOT=/host/sys \
  -e SPLATTY_SERVER_TOKEN=<token> \
  -e SPLATTY_URL=https://splatty.app \
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
  | sudo SPLATTY_SERVER_TOKEN=<token> sh
```

When `SPLATTY_SERVER_TOKEN` is set on a Linux host with systemd, the script also creates a
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

## Install (go)

Requires Go 1.25 or newer:

```bash
go install github.com/splatty-hq/splatty-agent@latest
```

The binary lands in `$(go env GOPATH)/bin`. This builds from source and sets up nothing
else — use the install script above if you want the systemd service too.

## Upgrade

Re-run the install script, then restart the service — the script replaces the binary but
never restarts a running agent, so the old process keeps serving until you do:

```bash
curl -fsSL https://raw.githubusercontent.com/splatty-hq/splatty-agent/master/install.sh | sudo sh
sudo systemctl restart splatty-agent
```

Leave `SPLATTY_SERVER_TOKEN` out when upgrading. With it set, the script rewrites
`/etc/splatty-agent.env` and the unit file, discarding anything you tuned by hand.

Set `VERSION=<tag>` to pin a release, or to roll back to an earlier one.

## Version

Release builds are stamped with their tag via `-ldflags -X main.version=<tag>`:

```bash
splatty-agent --version   # splatty-agent v1.0.0
```

Builds without that flag fall back to the module version the toolchain records — the
tag for `go install` binaries, the commit for local builds. For images, pass `--build-arg VERSION=<tag>` to `docker build`.
The running agent also logs its version on startup, so `journalctl -u splatty-agent` shows
which build is live.

## Run (systemd, manual)

If you'd rather not run the script, the unit file ships in the repo. Copy the binary to
`/usr/local/bin/splatty-agent` and `systemd/splatty-agent.service` to `/etc/systemd/system/`.
Create the `splatty` user (`useradd --system --no-create-home --shell /usr/sbin/nologin splatty`)
and write `/etc/splatty-agent.env`:

```
SPLATTY_SERVER_TOKEN=<token>
SPLATTY_URL=https://splatty.app
```

Then `systemctl enable --now splatty-agent`.

## License

[MIT](LICENSE)

## Metrics emitted

| Metric | Tags | Linux | macOS |
|---|---|:-:|:-:|
| `cpu.usage_percent` | — | ✅ `/proc/stat` | ✅ `top -l 2 -n 0 -s 1` |
| `mem.total_bytes` | — | ✅ `/proc/meminfo` | ✅ `hw.memsize` |
| `mem.used_bytes` | — | ✅ `/proc/meminfo` | ✅ `vm_stat` (active + wired + compressor) |
| `mem.available_bytes` | — | ✅ `/proc/meminfo` | ✅ `vm_stat` (free + inactive + speculative) |
| `swap.total_bytes`, `swap.free_bytes` | — | ✅ `/proc/meminfo` | ✅ `vm.swapusage` |
| `swap.used_bytes` | — | ✅ `SwapTotal - SwapFree` | ✅ `vm.swapusage` |
| `load.1`, `load.5`, `load.15` | — | ✅ `/proc/loadavg` | ✅ `vm.loadavg` |
| `disk.total_bytes`, `disk.used_bytes`, `disk.free_bytes` | `mount` | ✅ `statfs` | ✅ `statfs` |
| `disk.read_bytes_per_sec`, `disk.write_bytes_per_sec` | `device` | ✅ `/proc/diskstats` | ✅ `ioreg` |
| `disk.read_ops_per_sec`, `disk.write_ops_per_sec` | `device` | ✅ `/proc/diskstats` | ✅ `ioreg` |
| `net.rx_bytes_per_sec`, `net.tx_bytes_per_sec` | `iface` | ✅ `/proc/net/dev` | ✅ `NET_RT_IFLIST2` |
| `net.rx_packets_per_sec`, `net.tx_packets_per_sec` | `iface` | ✅ `/proc/net/dev` | ✅ `NET_RT_IFLIST2` |

On macOS, CPU + used/available memory aren't reachable through sysctl; the only sources are
the Mach `host_statistics` API (cgo-only) or the bundled tools. The agent shells out to
`top` (CPU sampling blocks for ~1s) and `vm_stat` instead, both of which ship with every
macOS install. Disk I/O counters sit behind IOKit, which needs cgo, so the agent reads the
same `IOBlockStorageDriver` statistics from `ioreg` and skips mounted disk images.
Network counters come from the `NET_RT_IFLIST2` routing sysctl (64-bit `if_data64`), limited
to the `en*` ports. `lo0`, `utun` VPN tunnels, `awdl`/`llw` and `bridge0` are skipped
because they would count the same traffic twice.

Disk I/O is reported per whole physical disk: partitions and virtual devices (`loop`,
`dm-*`, `md*`, `zram`) are skipped so stacked devices don't count the same I/O twice. In a
container, mount the host's `/sys` and set `SYSFS_ROOT` or the agent can't tell them apart
and reports no disk I/O.
