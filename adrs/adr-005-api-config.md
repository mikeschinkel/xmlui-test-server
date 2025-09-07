# ADR 005: XMLUI API Spec (v2) — Routes, Parameters, and Result Shaping

* **Status:** Accepted
* **Date:** 2025-09-03
* **Authors:** XMLUI team
* **Related:** ADR-0001 (overall config & single-DB), ADR-0002 (load/validate/migrate)

---

## 1) Context

We need a small, friendly JSON format to define HTTP → SQL endpoints for the XMLUI test server. Non-experts should be able to read and author files quickly, while the server enforces safe parameter binding and predictable result shapes. This ADR nails down the **v2 API spec** (the `api.json` schema) covering routes, parameters, and results.

---

## 2) Decision (v2 API Spec)

### 2.1 Versioning & schema

* Every `api.json` includes:

    * `"$schemaVersion": 2` (integer)
    * `"$schema": "https://xmlui-org.github.io/schemas/v2/api.schema.json"`
* The server treats this as **API spec v2**. (v1 compatibility/migration is defined elsewhere.)

### 2.2 File shape (high-level)

* **`webroot`**: the URL prefix under which endpoints mount (replaces `basePath`).
* **`endpoints[]`**: list of endpoint objects.
* Each endpoint:

    * `"path": "METHOD /route/with/{pathParams}"`  (exactly one method per endpoint)
    * exactly one of `sql` (inline string) **or** `sqlFile` (path to SQL)
    * optional `params` map (shorthand; see §2.4)
    * **result shaping** via `rows_expected` and `rowType` (see §2.5)

### 2.3 Route form

* `"path"` must match `^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\s+\/.*$`
* Path parameters use braces, e.g. `/users/{id}`, `/orders/{orderId}/items/{itemId}`

### 2.4 Parameters (shorthand map)

* Map key = param name; append `?` to mark **optional**.
* Map value = simple type: `"string" | "int" | "real" | "bool"`
* **Source inference**:

    * If `{name}` appears in the route → **path** parameter.
    * Otherwise → **query** parameter.
* **Binding**: SQL uses **named placeholders** `:name`. Every `:name` must appear in `params`, and every `{name}` in the route must have a matching `params` entry.
* Optional params (with `?`) may be omitted; their bound value is `NULL` in SQL.

### 2.5 Result shaping

* **`rows_expected`**: `"one"` or `"many"`

    * Append `?` to allow empty results: `"one?"` (0–1 row), `"many?"` (0..N rows).
    * Without `?`: `"one"` (exactly 1 row), `"many"` (≥1 row).
    * **Default** if omitted: `"many?"`.
* **`rowType`**: how to serialize each row:

    * `"json"`: row as an object (`{ columnName: value, ... }`) — **default**
    * `"int" | "real" | "string"`: use **first column only**, coerced to that scalar
    * `"columns"`: row as an array with auto typing, e.g. `[1,"Ada",99.5]`
    * `"columns(T1,T2,...)":` typed tuple; `Tn` ∈ `int | real | string | json`
      (the arity must equal the number of selected columns)
* **HTTP behavior (recommended defaults)**:

    * `rows_expected = "one"`: 0 rows → **404**; >1 row → **400** (contract error)
    * `rows_expected = "one?"`: 0 rows → **200 + null**
    * `rows_expected = "many"`: 0 rows → **404**
    * `rows_expected = "many?"`: 0 rows → **200 + \[]**

---

## 3) Examples

### 3.1 Minimal file

```json
{
  "$schema": "https://xmlui-org.github.io/schemas/v2/api.schema.json",
  "$schemaVersion": 2,
  "title": "XMLUI API Spec",
  "name": "xmlui-test-server-demo-api",
  "description": "Demo API for xmlui-test-server",
  "webroot": "/api",
  "endpoints": []
}
```

### 3.2 Single object by ID (optional hit)

```json
{
  "path": "GET /users/{id}",
  "sql": "select id, name from users where id = :id",
  "params": { "id": "int" },
  "rows_expected": "one?",
  "rowType": "json"
}
```

### 3.3 List with query params (may be empty)

```json
{
  "path": "GET /search",
  "sql": "select id, name from widgets where name like :q || '%' limit coalesce(:limit, 25)",
  "params": { "q": "string", "limit?": "int" },
  "rows_expected": "many?",
  "rowType": "json"
}
```

### 3.4 Scalar count (must exist)

```json
{
  "path": "GET /count",
  "sql": "select count(*) from widgets where name like :q || '%'",
  "params": { "q": "string" },
  "rows_expected": "one",
  "rowType": "int"
}
```

### 3.5 Typed tuples (must return at least one row)

```json
{
  "path": "GET /top",
  "sql": "select id, name, score from leaderboard order by score desc limit :n",
  "params": { "n?": "int" },
  "rows_expected": "many",
  "rowType": "columns(int,string,real)"
}
```

### 3.6 Using a SQL file

```json
{
  "path": "GET /query_from_file",
  "sqlFile": "sql/query.sql",
  "rows_expected": "many?",
  "rowType": "json"
}
```

---

## 4) Invariants (validation rules)

* File: `$schemaVersion === 2`; `$schema` points to `/v2/api.schema.json`.
* Endpoint: must have **exactly one** of `sql` or `sqlFile`.
* `path` must match the METHOD + route pattern and use `{name}` only for path params.
* **Parameters:**

    * For every `:name` placeholder in SQL, there is a `params.name` (or `params["name?"]`).
    * For every `{name}` in the route, there is a `params` entry for `name` (required unless you explicitly mark it optional as `name?`).
    * Param types: `string | int | real | bool`.
* **Result shaping:**

    * `rows_expected` ∈ `{ "one", "one?", "many", "many?" }` (default `"many?"`).
    * `rowType` ∈ `{ "json", "int", "real", "string", "columns", "columns(...)" }` (default `"json"`).
    * If `rowType = "columns(T1,...,Tn)"`, arity equals number of selected columns and each `Ti` is one of `int | real | string | json`.
    * Scalar `rowType` (int/real/string) is applied to the **first** selected column; extra columns are ignored.
* **Cardinality errors** follow the HTTP recommendations in §2.5 unless configured otherwise.

---

## 5) Backwards compatibility (v1 → v2, summary)

* `basePath` → `webroot`
* Per-endpoint `methods: { GET: {...}, ... }` → multiple endpoint entries with `"path": "METHOD /route"`
* Any older “result” notions map to:

    * `rows_expected`: `one`/`one?`/`many`/`many?`
    * `row_type`: `json`/`int`/`real`/`string`/`columns`
    * `column_types`: array of `json`/`int`/`real`/`string` for row_type=`columns`
* Param handling is stricter (named placeholders must match the `params` map).

*Migration is handled outside this ADR; see ADR-0002 for loader and migration strategy.*

---

## 6) Consequences

**Pros**

* Minimal cognitive load: METHOD/route path, tiny param map, two knobs for results.
* Safe parameter binding with named placeholders.
* Clear JSON outputs (object, scalar, or tuple) with predictable emptiness semantics.
* Stable evolution via `$schemaVersion` and versioned schema URLs.

**Cons**

* Scalar row types only bind the first column (by design).
* Typed `columns(...)` requires authors to keep SQL select list and signature aligned.

---

## 7) Follow-ups

* Publish `v2/api.schema.json` with these fields and validations.
* Document precise 404/200 behavior in server docs (and keep defaults here).
* Provide migration notes/examples for typical v1 specs in the repo’s docs.
