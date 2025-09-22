# ADR 007: APIParamsMap Marshaling and Unmarshalling

**Status:** Accepted
**Date:** 2025-09-22
**Authors:** Mike Schinkel [mike@newclarity.net](mailto:mike@newclarity.net)

---

## 1. Context

The XMLUI test server supports two representations of API parameters in configuration:

1. **Shorthand map form (`APIParamsMap`)**

   ```json
   {
     "id": "int",
     "slug": "slug:length[5..50]"
   }
   ```

2. **Explicit object array (`APIParamV1`)**

   ```json
   [
     { "name": "id", "type": "int", "constraints": "" },
     { "name": "slug", "type": "slug", "constraints": "length[5..50]" }
   ]
   ```

An endpoint (`APIEndpointV2`) may use either form. On load, the server tracks which form was used and round-trips back to the same form when marshaling.

This ADR documents the design and rules for marshaling/unmarshaling the `APIParamsMap` form.

---

## 2. Decision

Implement custom JSON marshaling/unmarshaling for `APIParamsMap` with the following semantics:

### 2.1 Structure

* **Top-level values allowed**:

    * `null` → treated as empty map on load; marshals back as `{}`.
    * `{}` → empty map.
    * Any other top-level type → error.

* **Member keys**:

    * Keys not starting with `@` → represent real parameters.
    * Keys starting with `@` → represent **comments**.

### 2.2 Member values

* **Real parameters** (keys without `@`):

    * Value **must be a JSON string**.
    * Strings are interpreted downstream as `<type>` or `<type>:<constraint>` (validation happens later in the PathVars compiler).
    * Any non-string value (`null`, bool, number, object, array) → error.

* **Comments** (keys with `@`):

    * Allowed as:

        * A single JSON string, or
        * An array of JSON strings (`["line1","line2"]`).
    * Any other type → error.
    * Array values are joined with newlines for storage, but round-trip as strings.

### 2.3 Ordering and round-trip

* Keys are stored and emitted in the order they appear in the JSON input.
* On marshal, the map is always written in preserved order.
* Whitespace is normalized out; round-trip uses canonical JSON spacing.
* Escapes may normalize (e.g., `\/` → `/`).

### 2.4 Duplicates

* Duplicate keys are not allowed.
* Encountering the same key a second time → `ErrAPIParamsMapDuplicateKey` with context:

    * `key`,
    * `new_value`,
    * `prev_value`,
    * `offset`.

### 2.5 Trailing data

* In `UnmarshalJSON([]byte)` (the byte-slice form), any **non-whitespace trailing data** after the object is an error.
* Error message includes both offset and a preview of trailing bytes.
* In `UnmarshalJSONFrom(*jsontext.Decoder)` (the streaming form), trailing detection is skipped (not possible in stream mode).

### 2.6 Error reporting

Errors use sentinel values with rich context joined via `errors.Join`:

* `ErrAPIParamsMapExpectedObject`
* `ErrAPIParamsMapStringsOnly`
* `ErrAPIParamsMapCannotBeNested`
* `ErrAPIParamsMapCannotContainArray`
* `ErrAPIParamsMapDuplicateKey`
* `ErrAPIParamsMapTrailingData`

All include `key`, `value`, and `offset` where applicable.

---

## 3. Rationale

### 3.1 Developer-friendly

* Clear error messages: include the key, the offending value, and the byte offset.
* Disallows silently ignored data (duplicates, trailing garbage).

### 3.2 Predictable round-tripping

* Order preservation ensures user configs are not scrambled.
* `null` → `{}` normalization prevents confusion and unifies empty semantics.
* Comments (`@...`) are preserved but not validated, enabling annotations.

### 3.3 Two-phase validation

* This layer enforces **structural correctness** only (objects, keys, strings).
* Semantic validation of `<type>` and `<constraint>` is deferred to the PathVars compiler (per \[PathVars README]).

---

## 4. Alternatives Considered

* **Preserve byte-for-byte value escapes**
  Rejected for MVP as overkill; values are always stored as decoded Go strings.

* **Allow any JSON type as values**
  Rejected to avoid confusion; parameters must always be specified as strings.

* **Forbid comment arrays**
  Rejected; multi-line comments are useful for documentation.

---

## 5. Consequences

### 5.1 Benefits

* Strong structural validation at load time.
* Consistent, developer-friendly error messages.
* Safe round-trip semantics (ordering preserved, comments preserved).
* Maintains backward compatibility with existing shorthand `params` maps.

### 5.2 Trade-offs

* Escapes are normalized (not byte-preserved).
* Comment arrays are joined with newlines internally, which may differ from original authoring style.

---

## 6. Related ADRs

* **ADR-005**: API Spec v2 — defines the `params` shorthand map that this ADR implements.
* **ADR-006 (JSON Versioning)**: defines the universal `version` property across schemas.

Neither ADR conflicts with this ADR:

* ADR-005 allows the shorthand `params` map but leaves its encoding rules unspecified — this ADR fills that gap.
* ADR-006 governs schema versioning, not parameter marshaling.

---

Do you want me to also draft example JSON fragments (valid and invalid) to embed at the end of this ADR for extra clarity?
