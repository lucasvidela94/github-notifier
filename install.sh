#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────
#  github-notifier  —  installer
# ─────────────────────────────────────────────

BINARY=github-notifier
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/github-notifier"
SERVICE_DIR="$HOME/.config/systemd/user"
SERVICE_FILE="$SERVICE_DIR/$BINARY.service"
ENV_FILE="$CONFIG_DIR/env"

# ── helpers ──────────────────────────────────

info()    { printf '\n  \033[1;34m--\033[0m  %s\n' "$*"; }
success() { printf '  \033[1;32mok\033[0m  %s\n'   "$*"; }
warn()    { printf '  \033[1;33m!!\033[0m  %s\n'   "$*"; }
die()     { printf '\n  \033[1;31merr\033[0m %s\n\n' "$*" >&2; exit 1; }

require() {
    command -v "$1" &>/dev/null || die "'$1' is required but not found. $2"
}

# ── detect package manager ───────────────────

install_sys_deps() {
    info "Installing system dependencies..."

    if command -v pacman &>/dev/null; then
        sudo pacman -S --needed --noconfirm gtk3 libappindicator-gtk3 libnotify
    elif command -v apt-get &>/dev/null; then
        sudo apt-get update -qq
        sudo apt-get install -y libgtk-3-dev libayatana-appindicator3-dev libnotify-bin
    elif command -v dnf &>/dev/null; then
        sudo dnf install -y gtk3-devel libayatana-appindicator-gtk3-devel libnotify
    elif command -v zypper &>/dev/null; then
        sudo zypper install -y gtk3-devel libappindicator3-devel libnotify-tools
    else
        warn "Unknown package manager. Install manually: gtk3, libappindicator3, libnotify"
        warn "Then re-run this script."
        exit 1
    fi

    success "System dependencies installed."
}

# ── check Go ─────────────────────────────────

check_go() {
    if ! command -v go &>/dev/null; then
        die "Go is not installed. Install it from https://go.dev/dl and re-run this script."
    fi

    local ver major minor
    ver=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
    major=$(echo "$ver" | cut -d. -f1)
    minor=$(echo "$ver" | cut -d. -f2)

    if [[ "$major" -lt 1 || ( "$major" -eq 1 && "$minor" -lt 21 ) ]]; then
        die "Go 1.21 or later is required (found go$ver). Update at https://go.dev/dl"
    fi

    success "Go $ver found."
}

# ── token setup ───────────────────────────────

setup_token() {
    mkdir -p "$CONFIG_DIR"
    chmod 700 "$CONFIG_DIR"

    # If a valid token already exists, ask before overwriting.
    if [[ -f "$ENV_FILE" ]] && grep -q "^GITHUB_TOKEN=ghp_" "$ENV_FILE"; then
        warn "A token is already configured in $ENV_FILE"
        read -rp "    Overwrite it? [y/N] " answer
        [[ "$answer" =~ ^[Yy]$ ]] || { success "Keeping existing token."; return; }
    fi

    printf '\n  GitHub personal access token needed.\n'
    printf '  Create one at: https://github.com/settings/tokens/new\n'
    printf '  Required scopes: notifications  (+ repo for private repos)\n\n'

    local token
    while true; do
        read -rsp "  Paste your token: " token
        printf '\n'
        [[ "$token" == ghp_* || "$token" == github_pat_* ]] && break
        warn "Token does not look valid (should start with ghp_ or github_pat_). Try again."
    done

    cat > "$ENV_FILE" <<EOF
GITHUB_TOKEN=$token
# POLL_INTERVAL_SECONDS=60
EOF
    chmod 600 "$ENV_FILE"
    success "Token saved to $ENV_FILE"
}

# ── build ─────────────────────────────────────

build() {
    info "Downloading Go modules..."
    go mod tidy

    info "Building $BINARY..."
    go build -ldflags="-s -w" -o "$BINARY" .
    success "Build complete."
}

# ── install binary ────────────────────────────

install_binary() {
    mkdir -p "$INSTALL_DIR"
    cp "$BINARY" "$INSTALL_DIR/$BINARY"
    chmod +x "$INSTALL_DIR/$BINARY"
    success "Binary installed to $INSTALL_DIR/$BINARY"

    # Warn if ~/.local/bin is not in PATH.
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH."
        warn "Add this to your ~/.bashrc:  export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
}

# ── systemd service ───────────────────────────

install_service() {
    mkdir -p "$SERVICE_DIR"

    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=GitHub PR Notifications Tray
After=graphical-session.target
PartOf=graphical-session.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/$BINARY
Restart=on-failure
RestartSec=5
EnvironmentFile=$ENV_FILE

[Install]
WantedBy=graphical-session.target
EOF

    systemctl --user daemon-reload
    systemctl --user enable --now "$BINARY.service"
    success "Service enabled and started."
}

# ── run tests ─────────────────────────────────

run_tests() {
    info "Running tests..."
    if go test ./internal/... -count=1 2>&1 | grep -E "^(ok|FAIL|---)" ; then
        success "All tests passed."
    else
        die "Tests failed. Aborting installation."
    fi
}

# ── summary ───────────────────────────────────

print_summary() {
    printf '\n'
    printf '  ┌─────────────────────────────────────────────┐\n'
    printf '  │           github-notifier installed          │\n'
    printf '  └─────────────────────────────────────────────┘\n\n'
    printf '  The tray icon will appear after your next login,\n'
    printf '  or right now since the service just started.\n\n'
    printf '  Useful commands:\n'
    printf '    systemctl --user status github-notifier   check status\n'
    printf '    journalctl --user -u github-notifier -f   live logs\n'
    printf '    systemctl --user restart github-notifier  restart\n\n'
}

# ── main ──────────────────────────────────────

main() {
    printf '\n  github-notifier installer\n'
    printf '  ─────────────────────────\n'

    install_sys_deps
    check_go
    run_tests
    setup_token
    build
    install_binary
    install_service
    print_summary
}

main "$@"
