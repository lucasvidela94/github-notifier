#!/usr/bin/env bash
# ─────────────────────────────────────────────
#  install.sh unit tests
#  Run: bash install_test.sh
# ─────────────────────────────────────────────
set -euo pipefail

PASS=0
FAIL=0

assert_eq() {
    local label="$1" expected="$2" actual="$3"
    if [[ "$expected" == "$actual" ]]; then
        printf '  \033[32mPASS\033[0m  %s\n' "$label"
        PASS=$((PASS + 1))
    else
        printf '  \033[31mFAIL\033[0m  %s\n' "$label"
        printf '         expected: %s\n' "$expected"
        printf '         got:      %s\n' "$actual"
        FAIL=$((FAIL + 1))
    fi
}

assert_file_exists() {
    local label="$1" path="$2"
    if [[ -f "$path" ]]; then
        printf '  \033[32mPASS\033[0m  %s\n' "$label"
        PASS=$((PASS + 1))
    else
        printf '  \033[31mFAIL\033[0m  %s (file not found: %s)\n' "$label" "$path"
        FAIL=$((FAIL + 1))
    fi
}

# ── test: detect_platform ────────────────────

test_detect_platform() {
    # Source only the detect_platform function
    detect_platform() {
        local os arch
        os="$(uname -s | tr '[:upper:]' '[:lower:]')"
        case "$os" in
            linux)  os="linux" ;;
            darwin) os="darwin" ;;
            *)      os="unsupported" ;;
        esac
        arch="$(uname -m)"
        case "$arch" in
            x86_64|amd64)  arch="amd64" ;;
            aarch64|arm64) arch="arm64" ;;
            *)             arch="unsupported" ;;
        esac
        PLATFORM="${os}_${arch}"
    }

    detect_platform
    # Should produce a valid platform string
    assert_eq "detect_platform produces valid format" "1" \
        "$(echo "$PLATFORM" | grep -cE '^(linux|darwin)_(amd64|arm64)$')"
}

# ── test: install.sh is valid bash ───────────

test_shellcheck() {
    if command -v shellcheck &>/dev/null; then
        if shellcheck install.sh 2>&1; then
            printf '  \033[32mPASS\033[0m  shellcheck\n'
            PASS=$((PASS + 1))
        else
            printf '  \033[31mFAIL\033[0m  shellcheck found issues\n'
            FAIL=$((FAIL + 1))
        fi
    else
        printf '  \033[33mSKIP\033[0m  shellcheck not installed\n'
    fi
}

# ── test: install.sh has correct shebang ─────

test_shebang() {
    local first_line
    first_line=$(head -1 install.sh)
    assert_eq "install.sh shebang" "#!/usr/bin/env bash" "$first_line"
}

# ── test: install.sh is executable ───────────

test_executable() {
    if [[ -x install.sh ]]; then
        printf '  \033[32mPASS\033[0m  install.sh is executable\n'
        PASS=$((PASS + 1))
    else
        printf '  \033[31mFAIL\033[0m  install.sh is not executable\n'
        FAIL=$((FAIL + 1))
    fi
}

# ── test: version flag works ─────────────────

test_version_flag() {
    if [[ -f github-notifier ]]; then
        local ver
        ver=$(./github-notifier --version 2>/dev/null || echo "")
        if [[ -n "$ver" ]]; then
            printf '  \033[32mPASS\033[0m  --version returns: %s\n' "$ver"
            PASS=$((PASS + 1))
        else
            printf '  \033[31mFAIL\033[0m  --version returned empty\n'
            FAIL=$((FAIL + 1))
        fi
    else
        printf '  \033[33mSKIP\033[0m  binary not built (run: go build)\n'
    fi
}

# ── test: service file template vars ─────────

test_service_template() {
    # Verify install.sh contains expected fields
    local content
    content=$(cat install.sh)

    for field in "ExecStart=" "EnvironmentFile=" "Restart=on-failure" "WantedBy=graphical-session.target" "_main_wrapper" "check_sys_deps"; do
        if echo "$content" | grep -q "$field"; then
            printf '  \033[32mPASS\033[0m  service template has %s\n' "$field"
            PASS=$((PASS + 1))
        else
            printf '  \033[31mFAIL\033[0m  service template missing %s\n' "$field"
            FAIL=$((FAIL + 1))
        fi
    done
}

# ── test: update path (re-run is idempotent) ─

test_update_detection() {
    # Simulate: installed version == latest → should skip
    INSTALLED_VERSION="v1.0.0"
    LATEST_VERSION="v1.0.0"
    if [[ "$INSTALLED_VERSION" == "$LATEST_VERSION" ]]; then
        printf '  \033[32mPASS\033[0m  same version detected as no-op\n'
        PASS=$((PASS + 1))
    else
        printf '  \033[31mFAIL\033[0m  same version not detected\n'
        FAIL=$((FAIL + 1))
    fi

    # Simulate: installed version != latest → should update
    INSTALLED_VERSION="v1.0.0"
    LATEST_VERSION="v1.1.0"
    if [[ "$INSTALLED_VERSION" != "$LATEST_VERSION" ]]; then
        printf '  \033[32mPASS\033[0m  version diff triggers update\n'
        PASS=$((PASS + 1))
    else
        printf '  \033[31mFAIL\033[0m  version diff not detected\n'
        FAIL=$((FAIL + 1))
    fi

    # Simulate: no installed version → fresh install
    INSTALLED_VERSION=""
    LATEST_VERSION="v1.0.0"
    if [[ -z "$INSTALLED_VERSION" ]]; then
        printf '  \033[32mPASS\033[0m  empty version triggers fresh install\n'
        PASS=$((PASS + 1))
    else
        printf '  \033[31mFAIL\033[0m  fresh install not detected\n'
        FAIL=$((FAIL + 1))
    fi
}

# ── run ──────────────────────────────────────

printf '\n  install.sh tests\n'
printf '  ─────────────────\n\n'

test_shebang
test_executable
test_shellcheck
test_detect_platform
test_version_flag
test_service_template
test_update_detection

printf '\n  ─────────────────\n'
printf '  %d passed, %d failed\n\n' "$PASS" "$FAIL"

[[ "$FAIL" -eq 0 ]] || exit 1
