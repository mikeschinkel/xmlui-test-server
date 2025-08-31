#!/bin/bash
#
# Common variables for xmlui-test-server build scripts
#

# Project settings
BIN_DIR="bin"
BINARY_NAME="xmlui-test-server"
BINARY_PATH="$BIN_DIR/$BINARY_NAME"
GO_SQLITE3_PATH="xmluisvr/sqlite3"

# SQLite settings
SQLITE_VERSION="3450200"
SQLITE_YEAR="2024"
SQLITE_CFLAGS="-DSQLITE_ENABLE_LOAD_EXTENSION -DSQLITE_ALLOW_LOAD_EXTENSION"

# Extension settings
EXTENSION_VERSION="v1.2.0"
STEAMPIPE_EXTENSION="steampipe_sqlite_github.so"

# Build directories
BUILD_DIR="xmlui-test-server-build"
SQLITE_INSTALL_DIR="${BUILD_DIR}/sqlite-install"

# Build flags
BUILD_TAGS="sqlite3_load_extension"

# Platform detection
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
    esac
    
    export OS ARCH
}

# Helper functions
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

error() {
    echo "[ERROR] $*" >&2
    exit 1
}

ensure_bin_dir() {
    if [[ ! -d "$BIN_DIR" ]]; then
        log "Creating $BIN_DIR directory..."
        mkdir -p "$BIN_DIR"
    fi
}

check_macos() {
    if [[ "$OS" != "darwin" ]]; then
        error "This operation is specifically for macOS"
    fi
}

check_macos_arm() {
    if [[ "$OS" != "darwin" || "$ARCH" != "arm64" ]]; then
        error "This operation is specifically for macOS ARM (Apple Silicon)"
    fi
}

check_macos_intel() {
    if [[ "$OS" != "darwin" || "$ARCH" != "amd64" ]]; then
        error "This operation is specifically for macOS Intel (x86_64)"
    fi
}

check_patched_sqlite3() {
    if [[ ! -d "$GO_SQLITE3_PATH" ]]; then
        error "Patched go-sqlite3 not found at $GO_SQLITE3_PATH"
    fi
}

# Initialize platform detection when sourced
detect_platform