# gotty-piko

[中文文档](README_CN.md) | English

An efficient terminal-based remote assistance tool that integrates gotty and piko services. Designed for remote assistance in complex network environments, avoiding the high bandwidth dependency of traditional remote desktop solutions while eliminating the need for complex network configurations and public IP addresses.

[gotty](https://github.com/friddle/gotty) (forked with custom modifications)
[piko](https://github.com/andydunstall/piko)

**Note:**
1. Windows solution use [goxrdp-piko](https://github.com/friddle/goxrdp-piko)

## Features

- One-shot binary: gotty + piko in one command
- Daemon mode (default): runs in background
- Auto-generated session ID: `{user}_{dir}_{random}`
- Basic Auth enabled by default with auto-generated credentials
- notify-send interception: desktop notifications pushed to browser toast
- Webhook forwarding: Feishu-compatible notification relay
- Static file browsing via `/files/` path
- Port proxy via `/port/{port}` path
- tmux integration for persistent sessions
- Cross-platform: Linux, macOS, Android

## Architecture

```
gottyp (client)
  ├── gotty (local web terminal)
  │     ├── /{session}/          → terminal web UI
  │     ├── /{session}/files/    → static file browser
  │     └── /{session}/port/{p}  → port proxy
  └── piko client → piko server → browser (via CDN)
```

## Quick Start

### Server Deployment

```yaml
# docker-compose.yaml
version: "3.8"
services:
  clauded:
    image: friddlecopper/clauded-port-forward:latest
    container_name: clauded
    environment:
      - PIKO_UPSTREAM_PORT=8022
    ports:
      - "80:80"
    restart: unless-stopped
```

```bash
docker-compose up -d
```

### Client Usage

#### Download

| Platform | Binary | Size |
|----------|--------|------|
| Linux amd64 | `gottyp-linux-amd64` | ~21 MB |
| Linux arm64 | `gottyp-linux-arm64` | ~20 MB |
| macOS Intel | `gottyp-darwin-amd64` | ~22 MB |
| macOS Apple Silicon | `gottyp-darwin-arm64` | ~21 MB |

```bash
# Download (replace with your platform)
wget https://github.com/friddle/gotty-piko/releases/latest/download/gottyp-linux-amd64 -O ./gottyp
# China mirror
wget https://ghproxy.com/https://github.com/friddle/gotty-piko/releases/latest/download/gottyp-linux-amd64 -O ./gottyp

chmod +x ./gottyp
```

#### Start

```bash
./gottyp
# or with options
./gottyp --session myserver
./gottyp --auth=false
./gottyp --static-index=/home/user --attach-port=3000
./gottyp --daemon=false
```

Output:
```
========================================
Remote URL: https://clauded.friddle.me/user_project_a1b2/
Username:   x7kq3m
Password:   p9w2nfc8h4
Port Proxy: https://clauded.friddle.me/user_project_a1b2/port/3000
Files:      https://clauded.friddle.me/user_project_a1b2/files/
========================================
```

## Access Methods

| URL | Description |
|-----|-------------|
| `https://clauded.friddle.me/{session}/` | Terminal web UI |
| `https://clauded.friddle.me/{session}/files/` | Static file browser |
| `https://clauded.friddle.me/{session}/port/{port}` | Port proxy |

## Configuration

### Client Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--session` | Session ID (piko endpoint + gotty path) | `{user}_{dir}_{random}` |
| `--remote` | Piko server URL | `https://clauded.friddle.me` |
| `--auth` | Enable Basic Authentication | `true` |
| `--auth-name` | Auth username | auto-generated |
| `--pass` | Auth password | auto-generated |
| `--terminal` | Terminal type (zsh, bash, sh, etc.) | auto-select |
| `--tmux` | Use tmux for persistent sessions | `true` |
| `--daemon` | Run as daemon (background) | `true` |
| `--pid-file` | PID file path | `/tmp/gottyp.pid` |
| `--enable-notify` | Intercept notify-send | `true` |
| `--notify-webhook` | Webhook URL (Feishu compatible) | disabled |
| `--static-index` | Directory for /files/ | current directory |
| `--attach-port` | Port for /port/ proxy | disabled |
| `--auto-exit` | Auto exit after 24h | `true` |

### Subcommands

```bash
gottyp tmux list        # List tmux sessions
gottyp tmux kill-all    # Kill all tmux sessions and gottyp daemons
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SESSION` | Session ID |
| `AUTH_NAME` | Auth username |
| `PASS` | Auth password |
| `REMOTE` | Piko server URL |
| `TERMINAL` | Terminal type |
| `DAEMON` | Daemon mode |
| `ENABLE_NOTIFY` | Notify interception |
| `NOTIFY_WEBHOOK` | Webhook URL |
| `STATIC_INDEX` | Static file directory |
| `ATTACH_PORT` | Port proxy target |
