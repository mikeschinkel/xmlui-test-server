# ADR-009 — URL Arrays & Rows Syntax for PathVars

**Status:** Accepted 2025‑09‑25

**Owner:** Mike Schinkel

**Motivation:** Provide an explicit, easy, and safe way to declare and ingest 1‑D lists and 2‑D rows from URLs (and JSON bodies) so that the SQL layer can expand them into placeholder lists/rows without ever allowing identifier injection.

---

## Decision

1. **Type notation:** Use **Go‑style prefix array notation** everywhere.

    * 1‑D list of scalars: `[]T` → e.g., `{ids:[]int}`
    * Fixed length (optional feature): `[N]T` → e.g., `{top3:[3]int}`
    * Row (tuple) type: `[ colType, ... ]` with optional names on columns: `[ name:Type, ... ]` → e.g., `{row:[user_id:uuid, org_id:int]}`
    * 2‑D (list of rows): `[][ ... ]` → e.g., `{rows:[][user_id:uuid, org_id:int]}`
    * Nested arrays inside rows allowed, e.g., `{rows:[][user_id:uuid, tags:[]string]}`.

2. **Constraints (array‑level):** Use colon before constraints and **commas** between constraints (existing convention).
   Supported initially: `count[min..max]`, `unique`.

    * Example: `{ids:[]int:count[1..100],unique}`

3. **Resolution order per variable:**

    1. **Path** (CSV for 1‑D only; single segment)
    2. **Query**
    3. **JSON body**
       *First found wins.* (Route‑level override MAY be added later.)

4. **Query encodings:**

    * **1‑D lists:**

        * Repeated keys: `?id=1&id=2&id=3`
        * CSV: `?ids=1,2,3`
        * If both present, merge in encounter order; trim whitespace; empty items error.
    * **2‑D rows:** **dotted numeric indices (base‑0, contiguous)**:
      `rows.0.user_id=u1&rows.0.org_id=orgA&rows.1.user_id=u2&rows.1.org_id=orgB`

        * Indices must be `0..N‑1`, no gaps, no duplicates.
        * For each index, all declared columns must appear exactly once.
    * **Fallback (advanced):** URL‑encoded JSON in one parameter is accepted (e.g., `rows=[ ["u1","orgA"], ["u2","orgB"] ]`), but **not recommended** in docs.

5. **JSON body encodings:**

    * 1‑D: arrays (preferred) or CSV strings; dotted path allowed (e.g., `{filters.ids:[]int}`).
    * 2‑D: array‑of‑arrays **or** array‑of‑objects. If names are declared in the row type, object keys must match exactly; otherwise positional.

6. **Validation & normalization:**

    * **1‑D:** validate element types; apply `unique` (first‑wins); enforce `count` on item count.
    * **2‑D:** enforce rectangularity; per‑column type validation; column presence per row; enforce `count` on row count.
    * **Empty arrays:** error by default unless the var is optional (e.g., `{ids?}`) or has an explicit default.
    * **Normalization outputs:** 1‑D → `[]T`; 2‑D → `[][]T` (or `[]struct{...}` when named columns are used downstream in Go code).

7. **Security invariants (non‑negotiable):**

    * Request data **never** becomes SQL identifiers (no column/table/alias/name substitution).
    * Names in row types (e.g., `user_id`) come only from **route config** and are used for **validation & error messages**; binding is positional.
    * SQL layer produces **value placeholders only** (e.g., `$1`, `?`, `@p1`).
    * Identifier‑style features (e.g., ORDER BY column or ASC/DESC) are out of scope; document CASE/whitelists in SQL as the recommended pattern.

8. **SQL expansion policy (informative, not part of URL syntax):**

    * 1‑D lists map to placeholder lists for `IN (...)`.
    * 2‑D rows map to `(VALUES ...)` constructs (for bulk insert or join/filter) with placeholders only.
    * No separate `:tupleN` directive is introduced.

---

## Rationale

* **Prefix arrays** (`[]T`, `[][ ... ]`) align with Go and many developers’ intuition for “container first,” and yield a **single parsing rule**: parse zero or more leading `[]`/`[N]`, then a base type (scalar or row).
* **Dotted numeric indices** avoid bracket quirks in URLs, are readable, and fit our dotted‑path model.
* **Base‑0, contiguous** indices minimize ambiguity and map directly to typical programmatic array handling.
* **Strict security** keeps the SQL layer dumb, portable, and safe.

---

## Grammar (author‑facing spec)

```
Variable     := '{' Name OptMulti OptOpt ':' Type OptConstraints '}'
OptMulti     := '*' | ε               # existing: multi-segment path capture (orthogonal to arrays)
OptOpt       := '?' [ Default ] | ε   # existing: optional / optional-with-default
Type         := ArrayPrefix* Base
ArrayPrefix  := '[]' | '[' Int ']'    # variable-length or fixed-length
Base         := Scalar | Row
Scalar      ::= 'int' | 'uuid' | 'string' | 'bool' | 'date' | 'timestamp' | 'any' | ...
Row          := '[' Col (',' Col)+ ']'
Col          := [ Name ':' ] Type     # allows nested arrays inside row columns
OptConstraints := ':' Constraint (',' Constraint)* | ε
Constraint   := 'count[' Int '..' Int ']' | 'unique' | OtherFutureSafe
```

**Notes**

* Canonical forms only: **no** `{ids[]}` or `{id:[]}` sugars in v1.
* `*` (multi‑segment) and `?` (optional/default) retain their current semantics and are **orthogonal** to arrays.
* Commas inside bracketed payloads (e.g., row types, enums) do not split constraints.

---

## Resolution (source precedence)

**Allowed sources:** Path, Query, JSON body.

**No-duplication rule:** A given var may be supplied by **one source only**. If the same var is present in more than one source (e.g., Path **and** Query), return **400 Bad Request** with error code `AMBIGUOUS_SOURCE`, naming the var and the conflicting sources. The server **must not** silently pick one.

**Presence checks (deterministic order, not precedence):**

1. Check Path
2. Check Query
3. Check JSON Body
   Use this order only to *detect* which sources are present; it does **not** imply fallback selection. **Body is considered only when the var is absent in both Path and Query.**

**Optional/defaults:** If a var is optional (`?`) and absent in all sources, use its default (if declared). If the SQL references the var and no value/default exists, raise `MISSING_REQUIRED_VAR`.

**Per-source encoding recap:**

* **Path:** 1‑D only; single‑segment CSV (e.g., `/users/10,20,30`). Multi‑segment `*` is orthogonal and **not** an array indicator.
* **Query (1‑D):** Repeated keys and/or CSV; both forms belong to the single Query source and are **merged** in encounter order; trim whitespace; empty items error.
* **Query (2‑D):** Dotted numeric indices (`rows.<n>.<col>`), base‑0, contiguous. All keys for a given var belong to the single Query source.
* **Body:** Dotted paths allowed. 2‑D can be array‑of‑arrays or array‑of‑objects per the declared row type.

---

## Examples

### 1‑D list via query

```
GET /users?{ids:[]int:count[1..100],unique}
# /users?ids=10,20&id=20&id=30  → ids = [10,20,30]
```

### 1‑D list via path CSV

```
GET /users/{ids:[]int}
# /users/10,20,30 → ids = [10,20,30]
```

### 2‑D rows via query (base‑0 contiguous dotted indices)

```
GET /membership?{rows:[][user_id:uuid, org_id:int]}
# /membership?rows.0.user_id=u1&rows.0.org_id=orgA&rows.1.user_id=u2&rows.1.org_id=orgB
# rows = [ ["u1", orgA], ["u2", orgB] ]
```

### 2‑D rows via body (large batch)

```
POST /events/bulk?{rows:[][user_id:uuid, kind:string, ts:timestamp]}
Body (array-of-arrays):
{ "rows": [ ["8a…","login","2025-09-25T10:00:00Z"], ["9b…","logout","2025-09-25T10:05:00Z"] ] }

Body (array-of-objects):
{ "rows": [ {"user_id":"8a…","kind":"login","ts":"2025-09-25T10:00:00Z"}, {"user_id":"9b…","kind":"logout","ts":"2025-09-25T10:05:00Z"} ] }
```

### Nested arrays inside a row

```
GET /users/tags?{rows:[][user_id:uuid, tags:[]string]}
# /users/tags?rows.0.user_id=u1&rows.0.tags.0=a&rows.0.tags.1=b
```

---

## Acceptance Tests (author‑level; illustrative)

1. **Merge repeats + CSV (1‑D):**

    * Input: `?ids=1,2&id=2&id=3` with `{ids:[]int:count[1..100],unique}`
    * Output: `[1,2,3]`

2. **Path CSV (1‑D):**

    * Input path: `/users/10,20,30` for `{ids:[]int}`
    * Output: `[10,20,30]`

3. **2‑D dotted indices (OK):**

    * Input: `rows.0.user_id=u1&rows.0.org_id=orgA&rows.1.user_id=u2&rows.1.org_id=orgB`
    * Decl: `{rows:[][user_id:uuid, org_id:int]}`
    * Output: `[[u1,orgA],[u2,orgB]]`

4. **2‑D index gap (ERROR):**

    * Input: `rows.0.user_id=u1&rows.0.org_id=orgA&rows.2.user_id=u3&rows.2.org_id=orgC`
    * Error: `rows indices must be contiguous base‑0 (missing index 1)`

5. **2‑D duplicate index (ERROR):**

    * Input: two `rows.1.user_id` groups
    * Error: `duplicate row index 1`

6. **2‑D missing column (ERROR):**

    * Input missing `rows.1.org_id`
    * Error: `rows[1].org_id is required`

7. **`unique` semantics (1‑D):**

    * Input: `?ids=1&id=2&ids=1` with `{ids:[]int:unique}`
    * Output: `[1,2]` (first‑wins)

8. **`count` bounds (1‑D):**

    * Input: `?ids=` (empty) with `{ids:[]int:count[1..10]}`
    * Error: `ids requires between 1 and 10 items`

9. **Body objects match names (2‑D):**

    * Decl: `{rows:[][user_id:uuid, role:string]}`
    * Body row `{ "role":"admin","user_id":"u1" }` is valid (order‑independent).
    * Extra field → error; missing field → error.

10. **Nested arrays in query (2‑D):**

    * Decl: `{rows:[][user_id:uuid, tags:[]string]}`
    * Input: `rows.0.user_id=u1&rows.0.tags.0=a&rows.0.tags.1=b`
    * Output: `[{user_id:u1, tags:[a,b]}]`

11. **Identifier injection guard:**

    * Any attempt to bind user input into SQL identifiers must fail at compile/validation time.
    * Placeholders expand to value params only.

12. **Fallback JSON param (advanced):**

    * Input: `rows=%5B%5B%22u1%22,%22orgA%22%5D,%5B%22u2%22,%22orgB%22%5D%5D`
    * Decl: `{rows:[][string,string]}`
    * Output: `[[u1,orgA],[u2,orgB]]` (accepted but not promoted in docs)

---

## Compatibility & Non‑Goals

* No PHP‑style `[]` query keys; no aligned parallel lists for 2‑D.
* No string‑keyed maps in query for v1 (future scope; body JSON is preferred for maps).
* No tuple directives in SQL; SQL expansion inferred from the shape (list vs rows).
* Multi‑segment `*` remains about **path capture**, not arrays.

---

## Implementation Notes

* Parser implements the grammar above; normalize to canonical forms (prefix arrays only).
* Resolver enforces **Path → Query → Body** precedence with short‑circuit.
* Query parser for 2‑D groups keys by `rows.<n>.<col>`; validates **base‑0 contiguous**.
* Array constraints apply after normalization and element validation.
* SQL binder deduplicates placeholders per standard rules but never inlines identifiers.

---

## Future Work (explicitly out of v1)

* Map types in query (e.g., `{rows:map[string][...]]}`) with string keys.
* Whitelisted identifier selection for ORDER BY / column‑sets.
* Additional constraints (`distinct-by[field]`, `sort`, etc.).
* Route‑level knobs for resolution order or empty‑array policy.
* Allow "declaring" user-defined value sets as type, e.g. add `foobar=[foo:int,bar:string]` to API config which enables URL var like `{foobars:[]foobar}`

---

## Appendix: Quick Reference (cheatsheet)
**Security:** values only; identifiers never bound from input.

| Shape                   | Declaration Syntax                                     | Query Syntax                                                                             |
|-------------------------|--------------------------------------------------------|------------------------------------------------------------------------------------------|
| **1‑D list**            | <nobr>`{ids:[]int:count[1..100],unique}`</nobr>        | `?ids=1,2&id=2&id=3`                                                                     |
| **2‑D&nbsp;rows**       | <nobr>`{rows:[][user_id:uuid, org_id:int]}`</nobr>     | <nobr>`?rows.0.user_id=u1&rows.0.org_id=orgA`<br>`&rows.1.user_id=u2&rows.1.org_id=orgB` |
| **Nested array column** | <nobr>`?{rows:[][user_id:uuid, tags:[]string]}`</nobr> | `?rows.0.tags.0=a&rows.0.tags.1=b`                                                       |
