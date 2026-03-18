#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────
#  github-notifier  —  installer / updater
#
#  curl -fsSL https://raw.githubusercontent.com/lucasvidela94/github-notifier/master/install.sh | bash
# ─────────────────────────────────────────────

REPO="lucasvidela94/github-notifier"
BINARY="github-notifier"
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

# ── detect arch ──────────────────────────────

detect_platform() {
    local os arch

    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux)  os="linux" ;;
        darwin) os="darwin" ;;
        *)      die "Unsupported OS: $os" ;;
    esac

    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *)             die "Unsupported architecture: $arch" ;;
    esac

    PLATFORM="${os}_${arch}"
}

# ── get latest version from GitHub ───────────

get_latest_version() {
    local url="https://api.github.com/repos/${REPO}/releases/latest"
    LATEST_VERSION=$(curl -fsSL "$url" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//')

    if [[ -z "$LATEST_VERSION" ]]; then
        die "Could not determine latest version. Check https://github.com/${REPO}/releases"
    fi
}

# ── check if update is needed ────────────────

check_installed_version() {
    INSTALLED_VERSION=""
    if [[ -x "$INSTALL_DIR/$BINARY" ]]; then
        INSTALLED_VERSION=$("$INSTALL_DIR/$BINARY" --version 2>/dev/null || echo "")
    fi
}

# ── install system deps ─────────────────────

install_sys_deps() {
    info "Installing system dependencies..."

    if command -v pacman &>/dev/null; then
        sudo pacman -S --needed --noconfirm gtk3 libappindicator-gtk3 libnotify </dev/tty
    elif command -v apt-get &>/dev/null; then
        sudo apt-get update -qq </dev/tty
        sudo apt-get install -y libgtk-3-dev libayatana-appindicator3-dev libnotify-bin </dev/tty
    elif command -v dnf &>/dev/null; then
        sudo dnf install -y gtk3-devel libayatana-appindicator-gtk3-devel libnotify </dev/tty
    elif command -v zypper &>/dev/null; then
        sudo zypper install -y gtk3-devel libappindicator3-devel libnotify-tools </dev/tty
    else
        warn "Unknown package manager. Install manually: gtk3, libappindicator3, libnotify"
        warn "Then re-run this script."
        exit 1
    fi

    success "System dependencies installed."
}

# ── download binary ──────────────────────────

download_binary() {
    local url="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${BINARY}_${PLATFORM}"
    local tmp
    tmp="$(mktemp)"

    info "Downloading $BINARY $LATEST_VERSION for $PLATFORM..."
    if ! curl -fsSL -o "$tmp" "$url"; then
        rm -f "$tmp"
        die "Download failed. Check that a release exists for $PLATFORM at:\n  $url"
    fi

    mkdir -p "$INSTALL_DIR"
    mv "$tmp" "$INSTALL_DIR/$BINARY"
    chmod +x "$INSTALL_DIR/$BINARY"
    success "Binary installed to $INSTALL_DIR/$BINARY"

    # Warn if ~/.local/bin is not in PATH.
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH."
        warn "Add this to your shell rc:  export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
}

# ── token setup ──────────────────────────────

setup_token() {
    mkdir -p "$CONFIG_DIR"
    chmod 700 "$CONFIG_DIR"

    # If a valid token already exists, keep it.
    if [[ -f "$ENV_FILE" ]] && grep -q "^GITHUB_TOKEN=ghp_\|^GITHUB_TOKEN=github_pat_" "$ENV_FILE"; then
        success "Token already configured in $ENV_FILE"
        return
    fi

    printf '\n  GitHub personal access token needed.\n'
    printf '  Create one at: https://github.com/settings/tokens/new\n'
    printf '  Required scopes: notifications  (+ repo for private repos)\n\n'

    local token
    while true; do
        read -rsp "  Paste your token: " token </dev/tty
        printf '\n'
        [[ "$token" == ghp_* || "$token" == github_pat_* ]] && break
        warn "Token does not look valid (should start with ghp_ or github_pat_). Try again."
    done

    printf '\n  GitHub username (for PR comment notifications).\n'
    local ghuser
    read -rp "  Your GitHub username: " ghuser </dev/tty
    [[ -z "$ghuser" ]] && die "GitHub username is required."

    cat > "$ENV_FILE" <<EOF
GITHUB_TOKEN=$token
GITHUB_USER=$ghuser
# POLL_INTERVAL_SECONDS=60
EOF
    chmod 600 "$ENV_FILE"
    success "Config saved to $ENV_FILE"
}

# ── systemd service ──────────────────────────

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

# ── restart on update ────────────────────────

restart_service() {
    if systemctl --user is-active --quiet "$BINARY.service" 2>/dev/null; then
        systemctl --user restart "$BINARY.service"
        success "Service restarted with new version."
    fi
}

# ── summary ──────────────────────────────────

print_summary() {
    local action="$1"
    printf '\n'
    printf '  ┌─────────────────────────────────────────────┐\n'
    printf '  │        github-notifier %-10s           │\n' "$action"
    printf '  └─────────────────────────────────────────────┘\n\n'
    printf '  Version: %s\n\n' "$LATEST_VERSION"
    printf '  Useful commands:\n'
    printf '    systemctl --user status github-notifier   check status\n'
    printf '    journalctl --user -u github-notifier -f   live logs\n'
    printf '    systemctl --user restart github-notifier  restart\n\n'
    printf '  To update later:\n'
    printf '    curl -fsSL https://raw.githubusercontent.com/%s/master/install.sh | bash\n\n' "$REPO"
}

# ── main ─────────────────────────────────────

main() {
    printf '\n  github-notifier installer\n'
    printf '  ─────────────────────────\n'

    detect_platform
    get_latest_version
    check_installed_version

    # If already at latest, skip.
    if [[ "$INSTALLED_VERSION" == "$LATEST_VERSION" ]]; then
        success "Already at latest version ($LATEST_VERSION). Nothing to do."
        exit 0
    fi

    if [[ -n "$INSTALLED_VERSION" ]]; then
        info "Updating $INSTALLED_VERSION -> $LATEST_VERSION"
    fi

    install_sys_deps
    download_binary

    if [[ -n "$INSTALLED_VERSION" ]]; then
        # Update: just replace binary and restart.
        restart_service
        print_summary "updated"
    else
        # Fresh install: configure token and service.
        setup_token
        install_service
        print_summary "installed"
    fi
}

main "$@"
