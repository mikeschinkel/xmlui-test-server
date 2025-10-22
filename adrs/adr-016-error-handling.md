# ADR-016: Error Handling with `doterr`

**Status**: Proposed
**Date**: 2025-10-17
**Owner**: Mike Schinkel
**Decision type**: Architecture / Cross-cutting Concern

---

## 1. Context

The application needs a consistent, composable, and stdlib-native way to:

* Attach **sentinel errors** (typed categories) to failures.
* Add **structured metadata** (key/value fields) at each call layer.
* **Compose** failures along the call stack using joined errors.
* **Combine** independent failures (e.g., validation).
* **Find** app-specific typed errors (e.g., RFC 9457) inside a joined tree.

Prior attempts (custom error types, third-party wrappers, `%w` ad-hoc wrapping) led to fragmentation, brittle string matching, or loss of interoperability.

---

## 2. Decision

Adopt the **`doterr`** package (copy-in; zero dependency) as the project’s **single** mechanism for structured error composition.

### Canonical Rules

1. **One entry per function**
   Each function creates **one** `doterr` entry (if needed) and **explicitly joins** it with the error it received/produced.

2. **Sentinels first, cause last**

   * In entry construction, **sentinels go first**, then metadata key/values, then optional trailing cause.
   * `New()` accepts an optional trailing cause: `doterr.New(ErrRepo, "key", val, cause)`.
   * `With` never accepts a cause; it's for enrichment only.

3. **One way, not many**

   * Prefer the canonical names (`New`, `With`, `Combine`, `Find`, `Meta`, `Errors`).
   * Aliases exist **only** for dot-import ergonomics (e.g., `NewErr`, `WithErr`, `CombineErrs`, `FindErr`).
   * `With` is for same-function enrichment only; do not overload it to join causes.

4. **Everything returns `error`**
   No exported concrete error types. All public functions return the builtin `error`.

5. **New vs. Combine**

   * Use **`New()`** with optional trailing cause for **composition** (build one entry with its cause in one call).
   * Use **`doterr.Combine`** for **combining independent** failures (validation, fan-out operations).
   * `errors.Join()` can still be used explicitly if preferred, but `New()` with trailing cause is more ergonomic.

---

## 3. Principles (Why)

* **Stdlib purity**: Leverages `errors.Join`, `errors.Is`, `errors.As`. No wrappers that break composability.
* **Typed categorization**: **Sentinel errors** (e.g., `ErrRepo`, `ErrDriver`, `ErrConstraint`) replace brittle string parsing.
* **Observability**: Each layer contributes metadata; trees are introspectable without parsing strings.
* **Interoperability**: Works with app-specific types (e.g., RFC 9457) by joining and later using `Find[T]`.
* **Simplicity**: Minimal API and explicit composition at call sites.

---

## 4. Design & Conventions

### 4.1 Sentinel Errors

Define sentinels at package level; use them **first** in entries:

```go
var (
    ErrDriver   = errors.New("driver")
    ErrRepo     = errors.New("repo")
    ErrService  = errors.New("service")
    ErrTemplate = errors.New("template")
    ErrConstraint = errors.New("constraint")
    ErrValidation = errors.New("validation")
)
```

### 4.2 Composing Errors (one entry per function)

```go
// innermost - with trailing cause
func readDriver(id int) error {
    cause := fmt.Errorf("connection reset by peer")
    return doterr.New(ErrDriver, "sql", "SELECT ... WHERE id=?", "param", id, cause)
}

// middle - passing error chain as trailing cause
func readRepo(id int) error {
    if err := readDriver(id); err != nil {
        return doterr.New(ErrRepo, "table", "users", err)
    }
    return nil
}

// outer - continuing the chain
func readService(id int) error {
    if err := readRepo(id); err != nil {
        return doterr.New(ErrService, "op", "GetUser", err)
    }
    return nil
}
```

> Dot-import style (optional): `import . "…/doterr"` then use `NewErr`, `WithErr`, etc.

### 4.3 Same-Function Enrichment

Use `With` to add fields within the same function. `With` never accepts or joins a cause. It merges into the rightmost `doterr` entry inside `err` when present; otherwise it joins a new entry (rightmost) without altering any existing causes or their order.

```go
err = doterr.With(err, "attempt", retryCount)
```

### 4.4 Combining Independent Errors

For multi-field validation or fan-out:

```go
var errs []error
for _, f := range fields {
    if bad(f) {
        errs = append(errs, doterr.New(ErrConstraint, "field", f))
    }
}
if ce := doterr.Combine(errs); ce != nil {
    return errors.Join(doterr.New(ErrValidation, "param", "payload"), ce)
}
```

### 4.5 Finding App-Specific Errors (e.g., RFC 9457)

`Find[T]` returns the first matching typed error via `errors.As`:

```go
if rfc, ok := doterr.Find[*rfc9457.Error](err); ok {
    // map joined metadata/sentinels into rfc as needed
}
```

### 4.6 Inspecting Metadata & Errors on an Entry

For a **single** `doterr` entry (both functions automatically unwrap one level to find the first doterr entry):

```go
if kvs := doterr.Meta(err); kvs != nil {
    for _, kv := range kvs {
        fmt.Printf("%s=%v\n", kv.Key, kv.Value)
    }
}
if errs := doterr.Errors(err); errs != nil {
    // errs are []error (sentinels or any error type); use errors.Is() for checks
    for _, e := range errs {
        if errors.Is(e, ErrRepo) {
            // handle repo error
        }
    }
}
```

(Deeper traversal across a whole tree is left to caller code using `Unwrap()`; we keep the API minimal.)

---

## 5. Usage Policy

* **Always** place **sentinels first** in `New/With` arguments, then metadata, then optional cause (for `New` only).
* **Exactly one entry per function layer** (avoid multiple `New` calls in a single function unless combining unrelated failures).
* **Use `New()` with trailing cause** for cross-function composition: `doterr.New(ErrRepo, "key", val, err)`.
* **Use `With` only for same-function enrichment**; `With` never accepts a cause.
* **Do not** pass a cause to `With`; `With` is enrichment-only.
* **Do not** export custom error types in this app unless absolutely necessary; prefer sentinels + metadata in entries.

---

## 6. Why No Type Introspection?

The API intentionally **does not** provide functions like `IsEntry(err) bool` or `IsCombined(err) bool` to detect internal types.

### Rationale

1. **Violates encapsulation**: The whole point of "everything returns `error`" is that consumers shouldn't care about concrete types. Exposing type checks undermines this principle.

2. **Use stdlib interfaces instead**: To detect multi-unwrappers (entries, combined, or stdlib joins), use the standard pattern:
   ```go
   if u, ok := err.(interface{ Unwrap() []error }); ok {
       // This is a multi-unwrapper (entry, combined, or errors.Join)
       for _, child := range u.Unwrap() {
           // traverse
       }
   }
   ```

3. **Existing API handles common cases**: `Meta()` and `Errors()` already unwrap one level automatically, covering most needs. For deeper traversal, use `Unwrap()` directly.

4. **Slippery slope**: Today it's "is this an entry?", tomorrow it's "give me the raw entry", then we've lost all abstraction benefits.

5. **Unclear use case**: Most operations (`errors.Is()`, `errors.As()`, `Meta()`, `Errors()`) work uniformly regardless of concrete type. If you need to distinguish types, you're probably doing something the API should handle for you.

### Historical Note

Prior art (like Go's stdlib hiding `joinError`) shows that keeping error structure opaque encourages robust, interface-based code. Type introspection leads to fragile coupling to implementation details.

---

## 7. Alternatives Considered

* **Custom error types per domain**
  Rejected: forces consumers to import/typeswitch, fragments error handling, and often breaks `errors.Is/As` usage.

* **Third-party error frameworks**
  Rejected: extra dependencies, non-stdlib semantics, higher cognitive load.

* **`fmt.Errorf("%w")` only**
  Rejected: single-cause wrapping obscures parallel context and makes metadata ad-hoc (string parsing risks).

* **Uber `multierr`**
  Useful library, but we prefer stdlib `errors.Join` + a minimal `Combine` to keep zero-dependency, copy-in philosophy.

---

## 8. Consequences

### Positive

* Uniform, readable error trees mirroring call stack.
* Safe, typed categorization (`errors.Is` with sentinels).
* Works with existing ecosystems (logging, tracing, RFC 9457 mapping).
* Minimal API; no dependency drag.

### Negative / Trade-offs

* Slight overhead when `With` merges into a join (linear in top-level child count).
  Mitigation: keep one entry per layer; depth is modest (typ. < 10).
* Callers must write their own traversal if they want whole-tree aggregation (intentional minimalism).

---

## 9. Migration Plan

1. Define sentinels for each package (driver/repo/service/validation/etc.).
2. Replace ad-hoc wrapping with `doterr.New(sentinel, …kvs…, cause)` (cause as last argument).
3. Replace multi-field validation accumulators with `doterr.Combine`.
4. For RFC 9457 surfaces, use `doterr.Find[*rfc9457.Error](err)` and map metadata/sentinels accordingly.
5. Update logging to print `err` plus (optionally) `doterr.Meta` values.

---

## 10. Testing Guidance

* **Table-driven tests** asserting:

   * `errors.Is(err, ErrX)` holds across layers.
   * Entry metadata is present and in order (`Meta`).
   * `With` enrichment preserves the **cause** and merges into the rightmost entry.
   * `Combine` returns nil/one/many correctly.
   * `Find[T]` pulls RFC 9457 / domain errors from joined trees.
   * `With` never reorders or replaces the rightmost cause in a joined tree.
   * `With(nil, kv...)` produces a single `doterr` entry (documented behavior).

* **Golden tests** (optional) for human-readable error strings.

---

## 11. Linting / CI

* Enforce "no exported function returns a concrete non-`error` type" (script/go-vet/go/types check).
* Optional: allow dot-import of `doterr` in linter config for test/example packages; normal import remains fine.
* Forbid calls where `With` receives a trailing error literal/identifier (enforce enrichment-only).
* Flag `With` used at cross-function boundaries (should use `New()` with trailing cause instead).
* Verify "sentinels before metadata before cause" in `New` argument lists.

---

## 12. Appendix: Patterns & Anti-patterns

**Do:**

```go
// Use New() with trailing cause for composition
return doterr.New(ErrRepo, "table", "users", err)

// Use With() for same-function enrichment only
err = doterr.With(err, "attempt", tries)

// Alternative: explicit errors.Join (more verbose but clearer in some cases)
entry := doterr.New(ErrRepo, "table", "users")
return errors.Join(entry, err)
```

**Don't:**

```go
// ❌ multiple New() calls in same function for the same chain
return errors.Join(doterr.New(ErrRepo, ...), doterr.New(ErrRepo, ...), err)

// ❌ string parsing for control flow - use errors.Is with sentinels
if strings.Contains(err.Error(), "repo") { ... }

// ❌ passing a cause to With (enrichment-only API)
err = doterr.With(err, "attempt", tries, someCause)

// ❌ using errors.Join when New() with trailing cause is simpler
entry := doterr.New(ErrRepo, "table", "users")
return errors.Join(entry, err) // just use: doterr.New(ErrRepo, "table", "users", err)
```

---

## 13. Open Questions

* Do we want optional formatters (e.g., JSON) as a separate tiny file later?
* Any sentinel naming standard to codify (prefix/es, package boundaries)?
* Where to house RFC 9457 mapping utilities (separate package vs. app code)?

---

**Outcome**: Upon approval, teams should adopt `doterr` for all new code and retrofit error handling in high-touch modules first (service, repo, validation).
