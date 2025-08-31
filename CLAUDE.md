# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xmlui-test-server is a lightweight Go HTTP server that provides:
- Static file serving from the current directory  
- `/query` endpoint for executing SQL against SQLite or PostgreSQL databases
- `/proxy` endpoint for proxying requests to external APIs (CORS bypass)
- Optional SQLite extension loading (e.g., Steampipe plugins)
- API endpoint routing based on JSON configuration files

## Build Commands

### Standard Build
```bash
go build -v
```

### Extension-Enabled Builds

**macOS ARM (Apple Silicon):**
```bash
./build-macos-arm.sh
```
Requires a locally patched go-sqlite3 repository at `$HOME/go-sqlite3` with the extension loading patch applied.

**Linux AMD64:**  
```bash
./build-linux-amd.sh
```
Downloads and builds a custom SQLite library with extension loading enabled.

## Running the Server

Basic usage:
```bash
./xmlui-test-server
```

Common options:
```bash
./xmlui-test-server --port 3000
./xmlui-test-server --extension steampipe-sqlite-github.so
./xmlui-test-server --api api.json --show-responses
./xmlui-test-server --pg-conn postgres://user@127.0.0.1:9193/db
```

## Architecture

### Core Components

**Server struct** (`Server`): Central handler containing database connection, API description, path regexes cache, and configuration flags.

**Database abstraction**: Supports both SQLite (default) and PostgreSQL with automatic parameter placeholder conversion (`?` → `$1`, `$2` for PostgreSQL).

**API Description system**: JSON-based configuration defining endpoints, HTTP methods, and SQL queries. Supports both inline SQL and external SQL files via `sqlFile` property.

**Parameter extraction**: Unified parameter handling from URL path segments (`:param`), query parameters, and JSON request bodies.

### Request Flow

1. **API requests** (`/api/*`): Route through API description matching, parameter extraction, SQL execution
2. **Direct queries** (`/query`): Accept POST requests with `{"sql": "...", "params": [...]}`  
3. **Proxy requests** (`/proxy/host.com/path`): Forward to `https://host.com/path` with CORS headers
4. **Static files** (`/*`): Serve from current directory, with `index.html` for root

### Database Integration

**SQLite**: Uses patched `go-sqlite3` with `C.sqlite3_enable_load_extension(db, 1)` for extension support. Sets `MaxOpenConns(1)` and `MaxIdleConns(1)` for thread safety.

**PostgreSQL**: Standard `lib/pq` driver with automatic parameter placeholder conversion.

**Extension loading**: SQLite extensions loaded via `SELECT load_extension(path)` with proper file permissions and error handling.

### API Description Format

JSON structure defining API endpoints:
- `basePath`: Common prefix for all endpoints  
- `endpoints[].path`: URL pattern with `:param` placeholders
- `methods`: HTTP method → SQL query mapping
- `sql` or `sqlFile`: Inline SQL or external file reference
- `params[]`: Named parameter binding order

Path parameters (`:id`) are extracted via regex and bound to SQL parameters in the order specified by the `params` array.