# ADR-004: Use `pg_url` for Postgres Connection String Variables

## Status
Accepted

## Context
We need a consistent, shorter alternative to `connection_string` for Postgres connection string variables and properties throughout the codebase.

## Decision
Use `pg_url` as the standard name for Postgres connection string variables.

## Rationale
- **Format alignment**: Postgres commonly uses `postgres://user:password@host:port/database` URL format
- **Postgres-specific**: The `pg_` prefix clearly identifies the database system
- **Common usage**: Most documentation and online examples use the URL format
- **Concise**: Much shorter than `connection_string` while remaining descriptive
- **Flexible**: Can accommodate both URL and key=value DSN formats if needed

## Consequences
- All new Postgres connection configuration should use `pg_url`
- Existing `connection_string` variables should be renamed to `pg_url` during refactoring
- Developers will immediately understand the variable purpose and format