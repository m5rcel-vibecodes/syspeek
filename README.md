# syspeek

A fast, beautiful, dependency-light system diagnostic tool written entirely in **Go**.

`syspeek` delivers a comprehensive overview of your machine in a single command. It combines the aesthetic clarity of `fastfetch`, the diagnostics of `htop`, the storage breakdown of `lsblk`, and the service awareness of `systemctl` into a unified, native Unix CLI tool.

```
┌─ SYSTEM ───────────────────────────────────┐
│ Host       m5rcel-mini                     │
│ OS         Arch Linux                      │
│ Kernel     6.18.9-zen1-2-zen               │
│ Arch       x86_64                          │
│ Uptime     3d 14h 21m                      │
│ Shell      zsh                             │
└───────────────────────────────────────────┘

HARDWARE

CPU        AMD Ryzen 5 5600
Cores      6 / 12
Usage      18%
Frequency  3.9 GHz

MEMORY

RAM        5.4 GB / 16.0 GB
Usage      34%

STORAGE

/          71% used [██████████████░░░░░░]  (142.0 GB / 200.0 GB)
/home      42% used [████████░░░░░░░░░░░░]  (210.0 GB / 500.0 GB)

NETWORK

Interface  enp6s0
IPv4       192.168.1.20
IPv6       2001:db8::1
RX         12.4 MB
TX         2.1 MB

SERVICES

✓ docker
✓ caddy
✓ ssh
✗ example.service
```

---

## Highlights

- **Ultra-Fast & Lightweight**: Sub-15ms startup time and minimal RAM usage.
- **Zero External Runtime Dependencies**: Implemented purely with the Go standard library and direct OS syscalls/virtual filesystems.
- **Deep Linux Support**: Direct `/proc`, `/sys`, `statfs`, and `/etc/os-release` parsing, hardware thermal sensors (`hwmon`, `thermal_zone`), init system detection (`systemd`, `OpenRC`, `sysvinit`), and GPU inspection (`nvidia-smi`, `drm`, `lspci`).
- **Native macOS (Darwin) Support**: Kernel metrics via `sysctl`, Mach virtual memory via `vm_stat`, `statfs` disk stats, `netstat` transfer accounting, and `launchctl` service integration.
- **Rock-Solid Error Handling**: Never panics or crashes when sensors, files, or services are unavailable; gracefully reports `N/A` or `Unsupported`.
- **Pure Machine-Readable JSON**: `--json` flag delivers structured data with zero escape sequences or formatting artifacts.
- **Dynamic Terminal Width**: Adapts table columns and box drawings to your terminal dimensions instead of wrapping awkwardly.
- **Standards Compliant**: Full compliance with the [NO_COLOR](https://no-color.org) specification.

---

## Installation

### From Source (Go 1.22+)

```bash
git clone github.com/m5rcel-vibecodes/syspeek.git
cd syspeek
go build -o syspeek ./cmd/syspeek
sudo mv syspeek /usr/local/bin/
```

Or using `go install`:

```bash
go install github.com/m5rcel-vibecodes/syspeek/cmd/syspeek@latest
```

### Using Make

```bash
make
sudo make install
```

---

## Command Reference

### Default Dashboard

```bash
syspeek
```

Renders the responsive terminal dashboard showing System, CPU, Memory, Storage, GPU, Network, and Services.

### Dedicated Subsystems

```bash
# Detailed CPU architecture, frequencies, load averages, usage gauge, thermals
syspeek cpu

# Detailed RAM (used, available, free, cached, buffers) and Swap statistics
syspeek memory

# Mounted filesystems, capacity, used/free space, and usage bars
syspeek disk

# Graphics accelerators, drivers, VRAM usage, and utilization
syspeek gpu

# Network interfaces, link state, IP addresses, MAC, and RX/TX statistics
syspeek network

# Service manager detection (systemd, OpenRC, launchd) and active/failed services
syspeek services

# Process monitor sorted by CPU and memory usage
syspeek top
syspeek processes
```

### JSON Output

Every command supports pure, machine-readable JSON output:

```bash
# Complete system snapshot as JSON
syspeek --json

# Subsystem JSON metrics
syspeek cpu --json
syspeek memory --json
syspeek disk --json
syspeek network --json
syspeek services --json
syspeek processes --json
```

### Live Refresh Mode (Watch)

Periodically refreshes metrics with clear-screen and smooth cursor restoration:

```bash
# Refresh dashboard every 2 seconds (default)
syspeek --watch

# Refresh dashboard every 1 second
syspeek --watch 1

# Refresh process table every 2 seconds
syspeek top
```

### Additional Flags

| Flag | Description |
|------|-------------|
| `--json` | Output pure JSON without ANSI colors or borders |
| `-w, --watch [N]` | Refresh interval in seconds (default: 2s) |
| `-c, --config <file>` | Path to custom configuration file |
| `--compact` | Compact layout with reduced vertical padding |
| `--no-color` | Disable colorized output (also honors `NO_COLOR=1`) |
| `-v, --version` | Print version information |
| `-h, --help` | Print help and usage guide |

---

## Configuration

`syspeek` optionally reads configuration from:

1. Path specified via `-c` / `--config`
2. Environment variable `$SYSPEEK_CONFIG`
3. `$XDG_CONFIG_HOME/syspeek/config.json`
4. `~/.config/syspeek/config.json`
5. `/etc/syspeek/config.json`

### Example `config.json`

```json
{
  "enabled_sections": [
    "system",
    "cpu",
    "memory",
    "storage",
    "network",
    "services"
  ],
  "refresh_interval": 2.0,
  "colors": true,
  "compact": false,
  "show_ascii_logo": false,
  "temperature_unit": "C",
  "network_interfaces": [
    "eth0",
    "wlan0"
  ],
  "disk_mounts": [
    "/",
    "/home"
  ],
  "service_filter": [
    "docker",
    "caddy",
    "sshd",
    "postgresql"
  ]
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled_sections` | `string[]` | `["system", "cpu", ...]` | Sections included in the default dashboard |
| `refresh_interval` | `float` | `2.0` | Refresh rate in seconds for `--watch` mode |
| `colors` | `bool` | `true` | Enable or disable ANSI terminal colors |
| `compact` | `bool` | `false` | Enable compact mode with reduced vertical padding |
| `show_ascii_logo` | `bool` | `false` | Show ASCII banner above the system box |
| `temperature_unit`| `string` | `"C"` | Temperature scale (`"C"` or `"F"`) |
| `network_interfaces`| `string[]` | `[]` | Priority interfaces (empty = auto-detect active) |
| `disk_mounts` | `string[]` | `[]` | Explicit mounts to display (empty = auto-detect) |
| `service_filter` | `string[]` | `[]` | Specific services to track |

---

## Cross-Compilation

`syspeek` has zero CGO requirements, making cross-compilation trivial:

```bash
# Linux x86_64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o syspeek-linux-amd64 ./cmd/syspeek

# Linux ARM64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o syspeek-linux-arm64 ./cmd/syspeek

# macOS Intel
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o syspeek-darwin-amd64 ./cmd/syspeek

# macOS Apple Silicon
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o syspeek-darwin-arm64 ./cmd/syspeek

# Windows x86_64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o syspeek-windows-amd64.exe ./cmd/syspeek
```

Or build all targets at once:

```bash
make cross-compile
```

---

## Docker

Build and run in a minimal container:

```bash
docker build -t syspeek .
docker run --rm -it \
  -v /proc:/proc:ro \
  -v /sys:/sys:ro \
  -v /etc:/etc:ro \
  syspeek
```

---

## Testing

Run unit tests:

```bash
go test -v ./...
```

Or via Make:

```bash
make test
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
