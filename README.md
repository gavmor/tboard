# tboard 📡

An interactive, privilege-free visual **Ping Dashboard TUI** built in Go using [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Bubbles](https://github.com/charmbracelet/bubbles).

## Features

- **Unicode Sparkline Graphs**: Real-time visualization using block runes (` `, `▂`, `▃`, `▄`, `▅`, `▆`, `▇`, `█`) color-coded dynamically by response latency.
- **Privilege-Free Probing**: Non-privileged TCP latency checks (defaulting to ports 80/443/53) requiring no `sudo` or raw socket root permissions.
- **Live Statistics**: Tracks Min RTT, Max RTT, Avg RTT, packet loss %, total packets sent/received, and last error diagnostics.
- **Interactive Controls**:
  - `j` / `k` or `↑` / `↓`: Navigate target list
  - `a`: Add new host dynamically (e.g. `example.com:80`)
  - `d`: Delete selected target
  - `Space`: Pause / resume latency checks
  - `r`: Reset statistics & history
  - `q` or `Ctrl+C`: Quit

## Prerequisites

- [mise](https://mise.jdx.dev/) toolchain manager (Go `1.26.6`)

## Getting Started

Run the TUI dashboard directly:

```bash
make run
```

Or execute via `mise`:

```bash
mise exec -- go run ./cmd/tboard
```

## Build & Test

```bash
make build    # Compiles executable to bin/tboard
make test     # Executes unit tests
```
