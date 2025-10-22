# ADR-017: NULL Handling in URL Templates for SQL

- **Status:** Accepted
- **Date:** 2025-10-19
- **Decision Owners:** XMLUI Core / Test-Server
- **Related:** ADR on parameter typing & constraints; ADR on SQL parameter binding

---

## 1. Context

XMLUI’s URL templating must capture request values that map cleanly to typed parameters and ultimately to SQL bindings. We need a way to:

1. Pass **SQL NULL** explicitly via the URL, without requiring URL-encoding.
2. Declare **nullability** and **default = NULL** within the template itself.
3. Keep parsing **single-pass** (no parser backtracking to mutate param attributes).
4. Remain unambiguous across paths and queries.

Existing template form (pre-change):

```
{<name>[?[<default>]]:<type>:<constraint>[,<constraint>]...}
```

It also supports an optional asterisk (`*`) after `<name>` as a multi-segment matching indicator, e.g. `/archive/{date*}` can match `/archive/2025/02/28`.

Values supported optionality (`?`), literal defaults, explicit typing, and constraints (e.g., `format[email]`). It lacked a standardized way to represent NULL in the URL and to express nullability and default NULL without complicating the parser.

---

## 2. Decision

We adopt **tilde (~)** as the explicit URL-level NULL sentinel and introduce a **type-prefix modifier** for nullability.

**A. URL NULL sentinel**

* `~` in a value position means **NULL**.
* Literal `~` is escaped by doubling: `~~` → literal `~`.

**B. Template nullability as a type modifier (no backtracking)**

* Attach `~` in front of the **type token** to mark the parameter as *nullable*:

    * `:~string`, `:~int`, etc.
* Keep constraints purely declarative; **do not** use a `nullable` constraint to set nullability, though we **may consider** adding it later for readability.

**C. Default NULL in templates**

* `?~` after the name indicates **default = NULL** (requires a nullable type).

If template authors forget to mark a type nullable (e.g., `{name?~:string}`), a **parse-time error** will be raised, reminding them to add the required `:~type`.

---

## 3. Specification

### 3.1 Grammar (delta)

```
<param>       ::= '{' <head> ':' <typed> [ ':' <constraints> ] '}'
<head>        ::= <name> [ '*' ] [ '?' [ <default> ] ]
<typed>       ::= [ '~' ] <type>
<constraints> ::= <constraint> (',' <constraint>)*
<default>     ::= <literal> | '~'        // '~' means default NULL
```

**Interpretation:**

* Leading `~` on **type** → the **value** is nullable.
* Optional `*` after name indicates multi-segment path matching.

### 3.2 URL Value Decoding

1. If the parameter is **missing** → state = *Missing*.
2. If present with raw value:

    * Exactly `~` → state = *Null*.
    * Otherwise, replace `~~` → `~` (single non-overlapping pass), then state = *Literal*.
3. Empty string in query (e.g., `?x=`) decodes to `""` (empty string), not NULL.

### 3.3 Binding & Validation

* If state = *Null*:

    * Allowed only when type is nullable (`:~type`).
    * Bind as **SQL NULL** (e.g., `nil` for Go `database/sql`).
    * Skip non-null constraints (format/regex/range); only nullability applies.
* If state = *Literal*:

    * Apply type conversion; reject `""` for non-string scalars (unless separately allowed by an `emptyok` capability).
    * Apply constraints (format, regex, enum, min/max, etc.).
* If state = *Missing*:

    * If a default exists:

        * `?~` → default NULL (must be `:~type`).
        * Otherwise → typed literal default.
    * Else if parameter is optional (`?` with no default):

        * Strings: default to `""`.
        * Non-strings: *Not Bound* (caller may omit from SQL bind list).
    * Else (required): 400 error.

---

## 4. Examples

**Optional string, default empty string (non-nullable)**

```
{q?:string}
```

* `?q=~` → 400
* `?q=`  → `""`
* omit   → `""`

**Optional string, nullable, default empty string**

```
{q?:~string}
```

* `?q=~` → `NULL`
* `?q=`  → `""`
* omit   → `""`

**Optional string, nullable, default NULL**

```
{q?~:~string}
```

* omit   → `NULL`
* `?q=~` → `NULL`
* `?q=`  → `""`

**Required int, nullable**

```
{age:~int}
```

* missing → 400
* `age=~` → `NULL`
* `age=18`→ `18`
* `age=`  → 400

**With constraints**

```
{email?~:~string:format[email]}
```

* `email=~` → `NULL` (skip format)
* `email=a@b.com` → pass format
* `email=a` → 400

**Multi-segment match**

```
/archive/{date*:~string}
```

* `/archive/~` → `date = NULL`
* `/archive/2025/02/28` → `date = ["2025","02","28"]` (semantics TBD)

---

## 5. Consequences

### 5.1 Benefits

* **Unambiguous NULL** without URL-encoding.
* **Single-pass parsing**: nullability discovered at the type token; constraints remain declarative.
* Works uniformly across **paths and queries**.
* Clear distinction between **missing**, **empty string**, and **NULL**.

### 5.2 Planned Future Enhancements

* Support for containers and nested nullability (e.g., `:~[~string]`) as a future feature.
* Consideration of a `nullable` constraint keyword for readability, which would be internally equivalent to using `:~type`.

### 5.3 Trade-offs

* `~` becomes a **reserved token**; requires doubling (`~~`) for literal tilde.
* Template authors must remember both `:~type` (nullable) and `?~` (default NULL).

---

## 6. Implementation Notes (Go)

* **Decode phase:** handle `~`/`~~` and missing/empty distinctions.
* **Typing phase:** verify nullable flag from the type token; enforce rules in §3.3.
* **SQL binding:** bind `nil` for NULL under `database/sql`.
* **Constraints:** bypass on NULL; apply otherwise.

---

## 7. Validation & Errors (RFC 9457 style)

* Non-nullable param given `~` → 400 (`invalid-null`).
* Empty string for non-string scalar → 400 (`invalid-empty`).
* Default NULL without nullable type → parse-time error (`invalid-default-null`).

---

## 8. Security Considerations

* Using a sentinel prevents overloading empty string as NULL, reducing coercion bugs.
* Always bind via parameters; never splice literals.
* Treat malformed escapes or unknown syntax as 400s to avoid ambiguity.

---

## 9. Alternatives Considered

* **Keyword `NULL`** in URLs: human-readable but conflicts with legitimate data.
* **Constraint `:nullable`**: acceptable, but requires parser backtracking.
* **Empty string as NULL**: conflates semantics.
* **Require URL-encoding**: user-hostile.

---

## 10. Summary

Adopt `~` as the cross-URL **NULL** sentinel (with `~~` escape). Express nullability via **type prefix** `:~type` and default NULL via **`?~`**. This preserves one-pass parsing, cleanly distinguishes missing/empty/NULL, and integrates with existing syntax including `*` for multi-segment path matching.
