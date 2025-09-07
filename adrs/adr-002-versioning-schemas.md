# ADR 0002: Load–Validate–Migrate Strategy for XMLUI JSON (Config, API, Extensions)

* **Status:** Accepted
* **Date:** 2025-09-01
* **Authors:** XMLUI team
* **Supersedes / Depends on:** Complements ADR-0001 (schemas & single-DB design)

---

## 1) Context

XMLUI’s test server reads user-authored JSON for three document types:

* **Test-server config** (single database, SQLite or Postgres)
* **API spec** (`api.json`; v1 legacy, v2 current)
* **SQLite extension descriptors** (optional)

We need a predictable runtime process to parse user files, validate them, and evolve formats over time without breaking users. This ADR defines that process and the behavior for schema versioning, autoloading extensions, and precedence/merge of multiple config sources.

---

## 2) Decision

Adopt a **Load → Validate (versioned) → Migrate (stepwise) → Validate (current)** pipeline for all user JSON, using:

* **Version markers:** `$schema` (URL to published JSON Schema), **`$schemaVersion` (integer)** matching the URL major (1, 2, …).
* **Version detection heuristics** when markers are absent (for `api.json` only).
* **Strict but small invariants** per version, enforced pre- and post-migration.
* **Source precedence** with last-one-wins and minimal merge semantics.
* **SQLite extension set-builder algorithm** (autoload + ordered overrides), not array merging.
* **CLI safety switch:** `--no-autoload` to disable extension directory scanning.

This isolates historical quirks inside versioned loaders and keeps the rest of the server working with one “current” in-memory model.

---

## 3) What “current” means (v2 snapshot)

* **API (v2):**
  `$schemaVersion: 2`, `$schema: https://xmlui-org.github.io/schemas/v2/api.schema.json`
  Uses `webroot`, `"path": "METHOD /route"`, `rows` + `rowType`, and param shorthand (`"id": "int"`, `"limit?": "int"`).
* **Config (v1):**
  `$schemaVersion: 1`, `$schema: https://xmlui-org.github.io/schemas/v1/test-server.schema.json`
  Single `database` with `type: "sqlite3" | "postgres"`. SQLite can declare `extensions` controls.
* **SQLite extension descriptor (v1):**
  `$schemaVersion: 1`, `$schema: https://xmlui-org.github.io/schemas/v1/sqlite3-extension.schema.json`.

---

## 4) Loader pipeline (applies to all doc types)

1. **Read bytes** (path is only a hint for error messages).
2. **Detect version:**

  * If `$schemaVersion` present → use that major.
  * Else **for API only**:

    * If `apiVersion` present or `methods` exists → treat as **v1**.
    * If endpoints use `"path": "METHOD /..."` or top-level `webroot` → treat as **v2**.
3. **Unmarshal** into that version’s struct(s).
4. **Validate (versioned invariants)** — fail fast with precise messages.
5. **Migrate stepwise** until “current” (e.g., v1→v2, v2→v3…); migrations are **pure** (no I/O).
6. **Validate (current invariants)** — guarantees the app always sees a clean model.
7. **Return current model** (or an error that cites detected version and offending field).

> Out of scope for this ADR: repository/package layout and language-level organization.

---

## 5) Source precedence & merge (config only)

Order (highest → lowest):

1. **CLI flags**
2. **`--config <file>`** (single optional file)
3. **Project file:** `./.xmlui/test-server.json`
4. **User file:** `~/.config/xmlui/test-server.json`
5. **Environment variables** (fill a field **only if** still unset)

Merge semantics:

* **Objects:** shallow replace at top-level keys (no deep merge).
* **Arrays:** whole-array replace.
* **Exception:** **SQLite extensions** are resolved via the set-builder algorithm in §7 (not by merging arrays).

---

## 6) API specifics (v1 → v2 migration rules)

* `basePath` → **`webroot`**.
* `methods{}` (v1) → **flatten** into one endpoint per method with `"path": "METHOD /route"`.
* Any v1 `result` notions → **`rows`** + **`rowType`** in v2:

  * `rows`: `"single" | "single?" | "multiple" | "multiple?"` (default `"multiple?"`).
  * `rowType`: `"object" | "int" | "float" | "string" | "columns" | "columns(T1,T2,...)"`
    (default `"object"`; scalar types use first column; `columns(...)` enforces arity/types).
* **Param shorthand**:

  * Required: `"id": "int"`, optional: `"limit?": "int"`.
  * If `{name}` appears in the route, it’s a **path** param; otherwise **query**.
  * SQL uses **named placeholders** `:name`; every `:name` must appear in `params`.

**Recommended empty-result HTTP defaults** (documented behavior):

* `single` with 0 rows → **404**; `single?` → **200** + `null`.
* `multiple` with 0 rows → **404**; `multiple?` → **200** + `[]`.

---

## 7) SQLite extensions (autoload + overrides)

**Autoload (default ON):** scan two **hardcoded** directories:

1. `./.xmlui/sqlite3/exts/`  (project-local, if exists)
2. `~/.config/xmlui/sqlite3/exts/`  (user-local, if exists)

Pick up both:

* `*.sqlite3-ext.json` **descriptors** (preferred when present)
* Raw binaries `*.so/.dylib/.dll` (**no descriptor required**)

**ID resolution:**

* Descriptor: use `id` if present; else filename stem.
* Raw binary: filename stem.
* Default entrypoint for raw binaries: `sqlite3_extension_init` (overridden by explicit entry where provided).

**Set-builder algorithm** (deterministic; replaces array merging):

1. Build the **autoload set** (id → loadable).
2. Apply each **`database.extensions`** item **in order**:

  * `{ "disable": "<id>" }` → remove by **id** (applies to autoloaded or previously added).
  * `{ "disable": "<path>" }` → remove by **path** (specific binary). Path can be:
    - **Absolute**: `/absolute/path/to/binary.so`
    - **Relative**: `./relative/path/binary.so` (relative to config file directory)
    - **Home-relative**: `~/sqlite3/exts/binary.so` (relative to user home)
    - **Security**: Parent directory references (`..`) are **forbidden** for security
  * `{ "id": "<id>" }` → ensure by **id** from autoload dirs (descriptor wins over raw if both exist).
  * `{ "descriptor": "<path-or-URL>" }` → load/replace by that descriptor's `id`.
  * `{ "id": "<id>", "file": "<bin>", "entry": "<symbol>", "sha256": "..." }`
    → explicitly load/replace a binary under that id; verify `sha256` if present.
3. Load the **final set** in a stable order (e.g., by id).

**CLI safety switch:** `--no-autoload` disables directory scanning; only explicit `extensions` items are considered.

---

## 8) Error reporting principles

* Always include **detected version** in error messages (e.g., `api.json (v1)`).
* Point to the **specific field** and give a short correction hint:

  * “`rows=single` forbids empty result; use `single?` to allow 0 rows.”
  * “SQL contains `:id` but `params.id` is missing.”
  * “Endpoint must specify exactly one of `sql` or `sqlFile`.”

---

## 9) Testing strategy (minimal set, high value)

1. **Version detection**

  * API v2 via `$schemaVersion: 2`; API v1 via `apiVersion`/`methods{}`; heuristic fallbacks work.
2. **Migration fixtures**

  * API v1→v2: `methods{}` flattening; `basePath`→`webroot`; `result`→`rows`/`rowType`.
3. **Param & SQL binding**

  * Path param inference from `{name}`; query param optional `?`; named placeholders coverage.
4. **Cardinality**

  * `single/single?/multiple/multiple?` behaviors and HTTP codes for empty results.
5. **Extensions**

  * Autoload on/off (`--no-autoload`); disable by id/path; explicit descriptor; explicit binary with `sha256`.
6. **Precedence & merge**

  * Flags override; whole-array replace; shallow key replace.

---

## 10) Schema publishing

* Host JSON Schemas at **`https://xmlui-org.github.io/schemas/`** in versioned folders:

  * `/v1/api.schema.json`, `/v1/test-server.schema.json`, `/v1/sqlite3-extension.schema.json`
  * `/v2/api.schema.json`
* Files should reference the **versioned** URL in `$schema`.
* `$schemaVersion` is an **integer** matching the URL major (1, 2, …).
* Optional: maintain a `/latest/` alias, but examples should keep version-pinned URLs.

---

## 11) Consequences

**Benefits**

* Small, testable surface area; future versions add only a new migration step.
* Users get “drop-in” success for raw SQLite3 binaries; descriptors add polish later.
* Clear, predictable overrides for extensions and minimal merge semantics for config.
* `$schema` + `$schemaVersion` provide editor help and stable evolution.

**Trade-offs**

* Single-DB server by design.
* Autoloading native binaries is inherently risky; this is **dev/test** tooling. `--no-autoload` offers a safety opt-out; explicit entries can pin by `sha256`.

---

## 12) Alternatives considered

* Deep merges / identity-based array merges → rejected (complexity, surprises).
* Profiles/overlays and multi-DB → rejected (scope creep vs first-success goal).
* Mandatory descriptors for extensions → rejected (hurts on-ramp).
* Embedding directory/package structure guidance here → out of scope.

---

