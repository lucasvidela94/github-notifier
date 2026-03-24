# github-notifier

[![Go Version](https://img.shields.io/github/go-mod/go-version/your-username/github-notifier)](https://github.com/your-username/github-notifier)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/your-username/github-notifier)](https://github.com/your-username/github-notifier/releases/latest)

A lightweight system tray application written in Go that surfaces GitHub pull request
comments as desktop notifications. Lives in your taskbar next to the volume and
bluetooth icons. Click any notification to open the exact comment in your browser
and mark it as resolved.

---

## Features

- System tray integration with bell icon
- Polls the GitHub Notifications API at a configurable interval
- Filters for pull request activity: comments, mentions, review requests
- Desktop notifications via `notify-send`
- Click-to-open comment links + mark as resolved
- SQLite local persistence
- Systemd user service with autostart

---

## Requirements

- Go 1.21+
- GTK3 + libayatana-appindicator3
- `notify-send`

Supported distros: Arch, Ubuntu/Debian, Fedora, openSUSE.

---

## Installation

```bash
git clone https://github.com/your-username/github-notifier
cd github-notifier
./install.sh
```

The installer handles everything:

1. Installs system dependencies
2. Runs the test suite
3. Prompts for GitHub token
4. Builds and installs the binary
5. Registers systemd user service

### GitHub Token

Create a token at `https://github.com/settings/tokens/new` with scopes:

| Scope | Purpose |
|-------|---------|
| `notifications` | Read and mark notifications |
| `repo` | Include private repo notifications |

Token is stored in `~/.config/github-notifier/env` with `600` permissions.

---

## Usage

### Systemd Service (recommended)

```bash
make enable        # Build, install, start
make logs          # Watch live logs
make disable       # Stop and disable
```

### Manually

```bash
export GITHUB_TOKEN=ghp_xxx
make run
```

---

## Tray Behavior

| Icon State | Meaning |
|------------|---------|
| Dark gray bell | No unread notifications |
| Orange bell + red dot | Unread notifications present |

### Activity Prefixes

| Prefix | Activity |
|--------|----------|
| `comment` | Someone commented on a PR |
| `mention` | You were mentioned with @handle |
| `team_mention` | Your team was mentioned |
| `review_requested` | You were asked to review |
| `author` | Activity on a PR you opened |

---

## Configuration

Environment variables in `~/.config/github-notifier/env`:

| Variable | Default | Description |
|----------|---------|-------------|
| `GITHUB_TOKEN` | required | Personal access token |
| `POLL_INTERVAL_SECONDS` | `60` | Poll frequency |
| `DB_PATH` | `~/.config/github-notifier/notifications.db` | SQLite path |

---

## Project Structure

```
github-notifier/
├── main.go                    # Entry point
├── icon/
│   └── icon.go                # Generated tray icons
├── internal/
│   ├── config/config.go       # Environment config
│   ├── db/db.go               # SQLite persistence
│   ├── github/client.go       # GitHub API client
│   ├── tray/tray.go           # Systray lifecycle
│   └── tray/ticker.go         # Poll interval
└── install.sh                 # Installer script
```

---

## Development

```bash
make test    # Run tests
make build   # Compile binary
make run     # Build and run
```

---

## Makefile Reference

| Target | Description |
|--------|-------------|
| `make build` | Compile binary |
| `make run` | Build and run |
| `make test` | Run tests |
| `make enable` | Install and start service |
| `make disable` | Stop and remove service |
| `make logs` | Tail service logs |
| `make clean` | Remove binary and service |

---

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) to keep our community
approachable and respectable.

---

## License

[MIT](LICENSE)

---

## Support

- Open an [Issue](https://github.com/your-username/github-notifier/issues)
- Check [SUPPORT.md](SUPPORT.md) for common problems
