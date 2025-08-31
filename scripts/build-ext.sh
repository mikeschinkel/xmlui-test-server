#!/bin/bash
#
# Build script for xmlui-test-server with extension support
# Usage: build-ext.sh [macos|linux]
#

set -e

# Get script directory and source variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/vars.sh"

usage() {
    echo "Usage: $0 [macos [arm|intel]|linux]"
    echo ""
    echo "Build xmlui-test-server with SQLite extension support for specified platform."
    echo "For macOS, specify architecture (arm for Apple Silicon, intel for x86_64)."
    echo "If no architecture specified for macOS, will auto-detect."
    exit 1
}

build_sqlite_linux() {
    log "Building SQLite with extension loading support for Linux..."
    
    mkdir -p "$BUILD_DIR"
    cd "$BUILD_DIR"
    
    # Download SQLite source if needed
    if [[ ! -f "sqlite-autoconf-$SQLITE_VERSION.tar.gz" ]]; then
        log "Downloading SQLite source..."
        wget "https://www.sqlite.org/$SQLITE_YEAR/sqlite-autoconf-$SQLITE_VERSION.tar.gz"
    fi
    
    # Extract if needed
    if [[ ! -d "sqlite-autoconf-$SQLITE_VERSION" ]]; then
        log "Extracting SQLite source..."
        tar xzf "sqlite-autoconf-$SQLITE_VERSION.tar.gz"
    fi
    
    # Build and install if needed
    if [[ ! -d "sqlite-install" ]]; then
        log "Configuring and building SQLite..."
        cd "sqlite-autoconf-$SQLITE_VERSION"
        CFLAGS="$SQLITE_CFLAGS" ./configure --prefix="../sqlite-install" --disable-shared --enable-static
        make CFLAGS="$SQLITE_CFLAGS" -j4
        make install
        cd ..
    fi
    
    cd ..
}

build_macos() {
    local arch="${1:-auto}"
    
    # Auto-detect architecture if not specified
    if [[ "$arch" == "auto" ]]; then
        if [[ "$ARCH" == "arm64" ]]; then
            arch="arm"
        elif [[ "$ARCH" == "amd64" ]]; then
            arch="intel"
        else
            error "Unsupported macOS architecture: $ARCH"
        fi
    fi
    
    log "Building $BINARY_NAME with extension loading support for macOS $arch..."
    
    check_macos
    check_patched_sqlite3
    ensure_bin_dir
    
    # Download extension for the specified architecture
    "$SCRIPT_DIR/download-ext.sh" macos "$arch"
    
    # Build with extension support from cmd directory
    CGO_ENABLED=1 \
    CGO_CFLAGS="$SQLITE_CFLAGS" \
    go build -tags "$BUILD_TAGS" -v -o "$BINARY_PATH" ./cmd
    
    chmod 755 "$STEAMPIPE_EXTENSION"
    log "Build complete! Run with: $BINARY_PATH --extension ./$STEAMPIPE_EXTENSION"
}

build_linux() {
    log "Building $BINARY_NAME with extension loading support for Linux AMD64..."
    
    ensure_bin_dir
    
    # Build custom SQLite
    build_sqlite_linux
    
    # Download extension
    "$SCRIPT_DIR/download-ext.sh" linux
    
    # Build with extension support from cmd directory
    CGO_ENABLED=1 \
    CGO_CFLAGS="-I$SQLITE_INSTALL_DIR/include $SQLITE_CFLAGS" \
    CGO_LDFLAGS="$SQLITE_INSTALL_DIR/lib/libsqlite3.a -lm -ldl" \
    go build -tags "$BUILD_TAGS" -v -o "$BINARY_PATH" ./cmd
    
    chmod 755 "$STEAMPIPE_EXTENSION"
    log "Build complete! Run with: $BINARY_PATH --extension ./$STEAMPIPE_EXTENSION"
}

main() {
    local platform="${1:-}"
    local arch="${2:-}"
    
    if [[ -z "$platform" ]]; then
        usage
    fi
    
    # Clean existing binary
    rm -f "$BINARY_PATH"
    
    case "$platform" in
        macos)
            # Validate architecture parameter for macOS
            if [[ -n "$arch" && "$arch" != "arm" && "$arch" != "intel" ]]; then
                echo "Error: Invalid macOS architecture '$arch'. Use 'arm' or 'intel'."
                usage
            fi
            build_macos "$arch"
            ;;
        linux)
            if [[ -n "$arch" ]]; then
                echo "Warning: Architecture parameter ignored for Linux (always uses amd64)"
            fi
            build_linux
            ;;
        *)
            echo "Error: Unknown platform '$platform'"
            usage
            ;;
    esac
}

# Run main function
main "$@"