# github-notifier

A lightweight system tray application written in Go that surfaces GitHub pull request
comments as desktop notifications. Lives in your taskbar next to the volume and
bluetooth icons. Click any notification to open the exact comment in your browser
and mark it as resolved.

---

## Features

- Sits in the system tray with a bell icon
- Polls the GitHub Notifications API at a configurable interval
- Filters exclusively for pull request activity: comments, mentions, review requests
- Sends a desktop notification (via `notify-send`) for each new event
- Clicking a tray entry opens the direct link to the comment and resolves it
- Persists state in a local SQLite database so nothing is lost between restarts
- Ships as a systemd user service that starts automatically with your graphical session

---

## Requirements

- Go 1.21 or later
- Arch Linux (or any distro with GTK3 and libayatana-appindicator)
- `notify-send` available in `$PATH` (provided by `libnotify`)

Install system dependencies on Arch:

```bash
make deps
```

This runs:

```bash
sudo pacman -S --needed gtk3 libappindicator-gtk3 libnotify
```

---

## GitHub token

Go to `https://github.com/settings/tokens/new` and create a classic personal
access token with the following scopes:

| Scope | Purpose |
|---|---|
| `notifications` | Read and mark notifications |
| `repo` | Include notifications from private repositories |

If you only work with public repositories, `notifications` alone is sufficient.

Store the token in the environment file:

```
~/.config/github-notifier/env
```

```ini
GITHUB_TOKEN=ghp_your_token_here
# POLL_INTERVAL_SECONDS=60
```

The file is created with permissions `600` so only your user can read it.

---

## Installation

```bash
# Clone
git clone https://github.com/you/github-notifier
cd github-notifier

# Set your token
nano ~/.config/github-notifier/env

# Build, install to ~/.local/bin, and enable the systemd user service
make enable
```

The service will now start automatically every time you log in.

---

## Usage

### As a systemd service (recommended)

```bash
# Enable and start
make enable

# Check status
systemctl --user status github-notifier

# Watch live logs
make logs

# Stop and disable
make disable
```

### Manually

```bash
export GITHUB_TOKEN=ghp_your_token_here
make run
```

---

## Tray behavior

| Icon state | Meaning |
|---|---|
| Dark gray bell | No unread notifications |
| Orange bell with red dot | One or more unread notifications |

The tray menu shows up to 15 entries. Each entry displays the PR title prefixed
by the activity type:

| Prefix | Activity |
|---|---|
| comment | Someone commented on a PR |
| mention | You were mentioned with @handle |
| team_mention | Your team was mentioned |
| review_requested | You were asked to review |
| author | Activity on a PR you opened |

Clicking an entry opens the comment directly in your browser and marks it as
resolved both locally and on GitHub.

---

## Configuration

All configuration is done through environment variables, typically set in
`~/.config/github-notifier/env`.

| Variable | Default | Description |
|---|---|---|
| `GITHUB_TOKEN` | required | Personal access token |
| `POLL_INTERVAL_SECONDS` | `60` | How often to check for new notifications |
| `DB_PATH` | `~/.config/github-notifier/notifications.db` | SQLite database path |

---

## Project structure

```
github-notifier/
├── main.go                        Entry point
├── icon/
│   └── icon.go                    Generates tray icons in memory (no asset files)
└── internal/
    ├── config/config.go           Reads environment variables
    ├── db/db.go                   SQLite persistence
    ├── github/client.go           GitHub API: fetch, filter, resolve
    └── tray/
        ├── tray.go                Systray lifecycle and menu
        └── ticker.go              Poll interval ticker
```

---

## Running tests

```bash
make test
```

Tests cover URL conversion, comment link resolution (with a mock HTTP server),
notification reason filtering, and all database operations.

```
ok  github-notifier/internal/db      (6 tests)
ok  github-notifier/internal/github  (5 tests)
```

---

## Makefile reference

| Target | Description |
|---|---|
| `make build` | Compile the binary |
| `make run` | Build and run directly |
| `make test` | Run all tests |
| `make enable` | Build, install, and start the systemd service |
| `make disable` | Stop and disable the service |
| `make logs` | Tail service logs with journalctl |
| `make install` | Install binary and autostart desktop entry |
| `make deps` | Install Arch system dependencies |
| `make clean` | Remove binary, service, and desktop entry |

---

## License

MIT
# github-notifier
