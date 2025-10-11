# ADR-012: Database Migrations Strategy


|Label| Status                                                                       |
|--|------------------------------------------------------------------------------|
|**Status:** | Pending Review                                                               |
|**Date:** | 2025-10-11                                                                   |
|**Author**: |ChatGPT _(after chat w/Mike Schinkel)_



## Context

XMLUI Test Server is the local development and testing backend for XMLUI, an ecosystem designed to make it easy for non-developers and AI assistants to build and deploy full web applications (including SaaS and internal apps). The Test Server provides a local HTTP/API service that integrates with SQLite initially, with planned support for PostgreSQL, MySQL/MariaDB, and DuckDB. It aims to abstract away complex database management tasks, such as schema diffs, migrations, and local test DB creation, while providing an approachable developer experience.

Previously, database schema management was ad hoc and manual. To improve developer productivity and ensure schema consistency across environments, we evaluated modern schema management tools written in Go that can be embedded within XMLUI’s server code.

Two main tools were evaluated:

* **Atlas (by Ariga)** — declarative and diff-based schema management, Apache 2.0 licensed, written in Go.
* **golang-migrate** — imperative migration runner, MIT licensed, also written in Go.

We also considered the need to support **DuckDB**, which is gaining traction for analytical workloads and small-scale local apps, but lacks robust external tooling for schema diffing and migrations.

---

## Decision

### Primary database management engine: **Atlas (ariga.io/atlas)**

The XMLUI Test Server will embed the **open-source Atlas engine** (Apache 2.0) for handling schema diffs, migrations, and linting. Atlas will serve as the default implementation for the internal `Migrator` and `Schemer` interfaces, providing cross-database schema management for supported backends (SQLite, PostgreSQL, MySQL/MariaDB, SQL Server).

### Atlas edition and licensing

* Use only the **open-source Atlas CLI and Go API**.
* Do **not** depend on Atlas Cloud or any commercial components.
* All individual XMLUI users (local developers) will have full functionality with the OSS version.
* Larger teams or PaaS customers can optionally integrate with Atlas Cloud, but this will not be required or bundled.

### Database coverage

| Database             | Role                                  | Support level                          |
| -------------------- | ------------------------------------- | -------------------------------------- |
| **SQLite**           | Default engine for local dev and test | ✅ Full Atlas support                   |
| **PostgreSQL**       | Primary production-like database      | ✅ Full Atlas support                   |
| **MySQL / MariaDB**  | Alternative production option         | ✅ Full Atlas support                   |
| **DuckDB**           | Analytical and local DB option        | ⚠️ Partial support (manual migrations) |
| **SQL Server, TiDB** | Optional future targets               | ✅ Supported by Atlas core              |

### DuckDB support strategy

Atlas does not yet officially support DuckDB because:

* DuckDB’s **DDL semantics are evolving**, with limited ALTER TABLE support.
* Its **SQL parser and serializer** lack full round-trip fidelity, which Atlas requires to diff and plan safely.

Therefore:

* DuckDB will be supported using a **simple SQL migration runner** that executes ordered `.sql` files (in the same style as `golang-migrate`).
* A future `DuckDBDialect` may be implemented for Atlas once the upstream API stabilizes.
* The XMLUI `Migrator` and `Schemer` abstractions will remain consistent across all engines, so a dialect can be added later without changing the external API.

### golang-migrate

`golang-migrate` was evaluated but rejected as the default because:

* It lacks declarative or diff-based capabilities.
* It requires users to hand-write migration files.
* It provides no linting or safety checks.

However, `golang-migrate` remains a fallback option if Atlas ever becomes unsuitable or if XMLUI users require the simplest possible migration runner for manual SQL workflows.

### Migration file conventions

* All migration files are stored under `db/migrations/<driver>/`.
* Filenames follow: `YYYYMMDDHHMMSS__short-description.sql`.
* Each driver may have its own subdirectory.
* Schema definitions (for declarative use) are stored under `db/schema/` in either HCL or SQL form.
* The migration history table defaults to `schema_migrations` (configurable).

### Internal architecture

A new package `internal/dbinfra` provides:

```go
type Migrator interface {
    Status(ctx context.Context) ([]AppliedMigration, error)
    Plan(ctx context.Context, from, to string) (Plan, error)
    Apply(ctx context.Context, upto string) error
}

type Schemer interface {
    Diff(ctx context.Context, desiredPath string) (Plan, error)
    ApplyPlan(ctx context.Context, p Plan) error
}

type Ephemeral interface {
    Start(ctx context.Context) (dsn string, stop func() error, err error)
}

type Console interface {
    ExecInteractive(ctx context.Context) error
}
```

Atlas provides the default implementation of `Migrator` and `Schemer`. Test databases for PostgreSQL and MySQL will use **testcontainers-go**, while SQLite and DuckDB will use in-memory or temporary file-based databases.

### Developer and API experience

Both **CLI** and **HTTP API** surfaces will expose database operations for human and AI-driven automation:

**CLI examples:**

```bash
xmlui dev db status
xmlui dev db diff
xmlui dev db apply
xmlui dev db console
xmlui dev db up --ephemeral --driver=postgres
```

**HTTP API examples:**

```
GET  /dev/db/status
GET  /dev/db/diff
POST /dev/db/apply
POST /dev/db/ephemeral
```

All endpoints will be disabled in production and only available when `--serve-dev` or `XMLUI_DEV=1` is set.

---

## Consequences

### Benefits

* **Low barrier to entry** for non-developers and AI agents — no manual SQL needed.
* **Cross-database abstraction** — consistent UX for SQLite, Postgres, and MySQL.
* **Safe, declarative workflows** — linting and diff validation reduce data-loss risk.
* **Future extensibility** — DuckDB and other engines can plug in later.
* **Embeddable** — Atlas Go API allows running migrations directly inside XMLUI, no external CLI required.
* **Offline friendly** — all open-source components, no SaaS dependencies.

### Drawbacks

* Atlas does not yet support DuckDB, requiring separate migration logic.
* Slightly higher learning curve for HCL schema format (if used).
* Dependency on an external open-source project with partial commercial backing (low risk, but worth tracking).

### Alternatives considered

* **golang-migrate:** simpler but requires SQL expertise.
* **sqldef family:** suitable for DuckDB or lightweight diffing, but less integrated for multi-DB support.
* **Manual SQL migrations:** feasible but error-prone and inconsistent across DBs.

---

## Future work

1. **Implement `xmluisvr/dbinfra/atlas.go`** adapter wrapping Atlas Go APIs.
2. **Implement `xmluisvr/dbinfra/duckdb.go`** simple runner for ordered SQL files.
3. Add `testcontainers-go` integration for PostgreSQL and MySQL ephemeral DBs.
4. Provide `/dev/db/ui` HTML interface for visualizing schema diff and migration status.
5. Consider future Atlas dialect for DuckDB once DDL stabilizes.
6. Reassess database strategy when XMLUI transitions from local-dev to hosted PaaS.

---

## References

* Atlas documentation: [https://atlasgo.io/docs](https://atlasgo.io/docs)
* Atlas Go API: [https://pkg.go.dev/ariga.io/atlas](https://pkg.go.dev/ariga.io/atlas)
* golang-migrate: [https://github.com/golang-migrate/migrate](https://github.com/golang-migrate/migrate)
* DuckDB documentation: [https://duckdb.org/docs/](https://duckdb.org/docs/)
* XMLUI Test Server: [https://github.com/mikeschinkel/xmlui-test-server](https://github.com/mikeschinkel/xmlui-test-server)
