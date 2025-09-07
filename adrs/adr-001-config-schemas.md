# ADR-001: Configuration Schema and Directory Structure

**Status:** Accepted  
**Date:** 2025-09-03  
**Authors:** XMLUI team

---

## 1. Context

XMLUI test server requires a consistent configuration system that can evolve over time while maintaining backward compatibility. The system must handle server configuration, API definitions, and database-specific settings across multiple file types and sources.

---

## 2. Decision

Adopt a **versioned schema system** with **standardized well-known directories** and **clear precedence rules** for configuration management.

### 2.1 Schema Versioning Strategy

**Version Markers**: All configuration files MUST include:
- `$schema`: URL pointing to the specific versioned JSON Schema
- `$schemaVersion`: Integer matching the schema major version

**Example**:
```json
{
  "$schema": "https://xmlui-org.github.io/schemas/v1/test-server.schema.json",
  "$schemaVersion": 1,
  "server": { ... },
  "database": { ... }
}
```

### 2.2 Well-Known Directory Structure

**Standard Directories** (searched in precedence order):
1. **Project-local**: `./.xmlui/` (highest precedence)
2. **User-local**: `~/.config/xmlui/` (lower precedence)

**Subdirectories**:
- `test-server.json` → Server configuration
- `api/` → API definition files  
- `sqlite3/exts/` → SQLite extension binaries and manifests
- `schemas/` → Local schema overrides (optional)

**Security**: Parent directory references (`..`) are forbidden in all path configurations.

### 2.3 Configuration File Types

| File Type | Schema URL | Purpose |
|-----------|------------|---------|
| **Server Config** | `/v1/test-server.schema.json` | Server settings, database config, extension control |
| **API Definition** | `/v2/api.schema.json` | API endpoints, SQL queries, parameter binding |
| **Extension Manifest** | `/v1/sqlite3-extension.schema.json` | SQLite extension metadata |

### 2.4 Configuration Precedence

**Source Precedence** (highest → lowest):
1. **CLI flags** (e.g., `--port`, `--config`)
2. **Explicit config file** (`--config path/to/file.json`)
3. **Project config** (`./.xmlui/test-server.json`)
4. **User config** (`~/.config/xmlui/test-server.json`)
5. **Environment variables** (fill unset fields only)

**Merge Semantics**:
- **Objects**: Shallow merge at top-level keys
- **Arrays**: Complete replacement (no merging)
- **Primitives**: Last-wins

---

## 3. Directory Resolution Rules

### 3.1 Path Types Supported

**Absolute paths**: `/absolute/path/to/file.json`
**Relative paths**: `./relative/path/file.json` (relative to config file directory)
**Home-relative paths**: `~/config/file.json` (relative to user home)

### 3.2 Extension Discovery

**SQLite Extensions** are discovered from:
1. `./.xmlui/sqlite3/exts/` (project extensions)
2. `~/.config/xmlui/sqlite3/exts/` (user extensions)

**File Types**:
- `*.sqlite3-ext.json` → Extension manifests (preferred)
- `*.so`, `*.dylib`, `*.dll` → Raw extension binaries

---

## 4. Schema Evolution Strategy

### 4.1 Backward Compatibility

**Within Major Version**: All changes MUST be backward compatible:
- New optional properties allowed
- New enum values allowed  
- Loosened validation rules allowed
- **Breaking changes require major version bump**

### 4.2 Migration Path

**Version Detection**: Use `$schemaVersion` field or heuristics for legacy files
**Migration Pipeline**: Load → Detect Version → Validate (versioned) → Migrate (stepwise) → Validate (current)
**Error Reporting**: Include detected version in all error messages

### 4.3 Schema Publishing

**Schema URLs**: `https://xmlui-org.github.io/schemas/v{major}/`
**Versioning**: Each major version gets its own directory
**Stability**: Published schemas are immutable

---

## 5. Configuration Structure

### 5.1 Root Configuration Format

```json
{
  "$schema": "https://xmlui-org.github.io/schemas/v1/test-server.schema.json",
  "$schemaVersion": 1,
  
  "server": {
    "port": 8080,
    "host": "localhost",
    "api_config": "./api/endpoints.json"
  },
  
  "database": {
    "type": "sqlite3 | postgres | duckdb | ...",
    "...": "type-specific configuration"
  }
}
```

### 5.2 Extensible Database Types

**Open-ended type system**: Database `type` field accepts any string value
**Conditional schemas**: Type-specific properties validated based on `type` value
**Future-proof**: New database types can be added without schema changes

**Examples**:
```json
{"type": "sqlite3", "filepath": "...", "extensions": [...]}
{"type": "postgres", "connection_string": "..."}
{"type": "mongodb", "connection_uri": "...", "collection_defaults": {...}}
```

---

## 6. Security Considerations

### 6.1 Path Security

- **Parent traversal forbidden**: Reject any path containing `..`
- **Absolute path validation**: Ensure absolute paths are within allowed directories
- **Home directory expansion**: Safely expand `~` to user home directory

### 6.2 Configuration Validation

- **Schema enforcement**: All configurations MUST validate against their declared schema
- **Required field validation**: Missing required fields cause startup failure
- **Type safety**: Strict type checking for all primitive values

---

## 7. Implementation Requirements

### 7.1 Configuration Loader MUST

- Support all path types (absolute, relative, home-relative)
- Implement proper precedence ordering
- Validate against versioned schemas
- Provide clear error messages with detected versions
- Reject configurations with parent directory references

### 7.2 Schema Design MUST

- Use `additionalProperties: true` strategically for extensibility
- Define clear required vs optional properties
- Use conditional schemas (`if/then`) for type-specific configuration
- Maintain backward compatibility within major versions

---

## 8. Examples

### 8.1 Minimal Configuration

```json
{
  "$schemaVersion": 1,
  "database": {
    "type": "sqlite3",
    "filepath": "test.db"
  }
}
```

### 8.2 Full Configuration with Extensions

```json
{
  "$schema": "https://xmlui-org.github.io/schemas/v1/test-server.schema.json",
  "$schemaVersion": 1,
  
  "server": {
    "port": 3000,
    "api_config": "api/endpoints.json",
    "show_responses": true
  },
  
  "database": {
    "type": "sqlite3",
    "filepath": "data/app.db",
    "open_params": {
      "_allow_load_extension": "1"
    },
    "init_sql": [
      "PRAGMA journal_mode=WAL;",
      "PRAGMA foreign_keys=ON;"
    ],
    "extensions": [
      {"id": "fts5"},
      {"disable": "legacy-extension"},
      {
        "id": "steampipe-github",
        "version": "v1.5.0",
        "download_urls": ["https://github.com/..."],
        "sha256s": {"darwin_arm64": "sha256:..."}
      }
    ]
  }
}
```

---

## 9. Consequences

### 9.1 Benefits

- **Single source of truth** for directory structure and precedence
- **Extensible system** that can accommodate new database types and features  
- **Version-aware evolution** with clear migration paths
- **Security** through path validation and schema enforcement
- **Developer experience** via clear error messages and schema validation

### 9.2 Trade-offs

- **Additional complexity** in configuration loading logic
- **Schema maintenance overhead** for version management
- **Migration complexity** for major version upgrades

---

## 10. Related ADRs

- **ADR-002**: Load-Validate-Migrate strategy (builds on this foundation)
- **ADR-003**: SQLite Extensions (uses well-known directories defined here)
- **ADR-004**: PostgreSQL URL naming (uses database type system defined here)