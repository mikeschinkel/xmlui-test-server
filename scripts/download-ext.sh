#!/bin/bash
#
# Download script for Steampipe SQLite extensions
# Usage: download-ext.sh [macos|linux]
#

set -e

# Get script directory and source variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/shared.sh"

usage() {
    echo "Usage: $0 [macos [arm|intel]|linux]"
    echo ""
    echo "Download Steampipe SQLite extension for specified platform."
    echo "For macOS, specify architecture (arm for Apple Silicon, intel for x86_64)."
    echo "If no architecture specified for macOS, will auto-detect."
    exit 1
}

download_extension() {
    local platform="$1"
    local arch="${2:-auto}"
    
    # Skip if extension already exists
    if [[ -f "$STEAMPIPE_EXTENSION" ]]; then
        log "Extension $STEAMPIPE_EXTENSION already exists, skipping download"
        return 0
    fi
    
    # Auto-detect architecture for macOS if not specified
    if [[ "$platform" == "macos" && "$arch" == "auto" ]]; then
        if [[ "$ARCH" == "arm64" ]]; then
            arch="arm"
        elif [[ "$ARCH" == "amd64" ]]; then
            arch="intel"
        else
            error "Unsupported macOS architecture: $ARCH"
        fi
    fi
    
    log "Downloading Steampipe SQLite extension for $platform $([ "$platform" == "macos" ] && echo "$arch")..."
    
    local platform_suffix
    case "$platform" in
        macos)
            case "$arch" in
                arm)
                    platform_suffix="darwin_arm64"
                    ;;
                intel)
                    platform_suffix="darwin_amd64"
                    ;;
                *)
                    error "Invalid macOS architecture: $arch"
                    ;;
            esac
            ;;
        linux)
            platform_suffix="${OS}_${ARCH}"
            ;;
        *)
            error "Unknown platform: $platform"
            ;;
    esac
    
    local url="https://github.com/turbot/steampipe-plugin-github/releases/download/$EXTENSION_VERSION/steampipe_sqlite_github.${platform_suffix}.tar.gz"
    log "Downloading from: $url"
    
    # Download and extract
    curl -L "$url" > ext.tar.gz
    tar xzf ext.tar.gz
    
    # Set permissions
    chmod 755 "$STEAMPIPE_EXTENSION"
    
    # Verify the file
    log "Extension downloaded and verified:"
    if command -v file >/dev/null 2>&1; then
        file "$STEAMPIPE_EXTENSION"
    else
        ls -la "$STEAMPIPE_EXTENSION"
    fi
    
    # Clean up
    rm -f ext.tar.gz
}

main() {
    local platform="${1:-}"
    local arch="${2:-}"
    
    if [[ -z "$platform" ]]; then
        usage
    fi
    
    case "$platform" in
        macos)
            # Validate architecture parameter for macOS
            if [[ -n "$arch" && "$arch" != "arm" && "$arch" != "intel" ]]; then
                echo "Error: Invalid macOS architecture '$arch'. Use 'arm' or 'intel'."
                usage
            fi
            download_extension "$platform" "$arch"
            ;;
        linux)
            if [[ -n "$arch" ]]; then
                echo "Warning: Architecture parameter ignored for Linux (always uses amd64)"
            fi
            download_extension "$platform"
            ;;
        *)
            echo "Error: Unknown platform '$platform'"
            usage
            ;;
    esac
}

# Run main function
main "$@"