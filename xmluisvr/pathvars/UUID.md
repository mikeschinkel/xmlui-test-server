# UUID / GUID Validation & Libraries (Go-centric)

This document summarizes algorithms, UUID versions, and Go libraries relevant to implementing a validator for path/query variables of the form:

```
{<varname>:uuid:format[<uuid_format>]}
```

where `<uuid_format>` might be `v4`, `v7`, `ulid`, `ksuid`, etc.

---

## UUID Versions (RFC 4122 / RFC 9562)

* **v1** – Time + MAC address (sortable but leaks MAC).
* **v2** – Time + POSIX UID/GID (rare).
* **v3** – Name-based, MD5.
* **v4** – Random (most common, Microsoft GUID style).
* **v5** – Name-based, SHA-1.
* **v6** – Reordered v1 (time-ordered, DB-friendly).
* **v7** – Unix timestamp (ms) + random (DB-friendly, modern default).
* **v8** – Custom / experimental fields.

RFC 9562 (2024) standardized **v6–v8**.

---

## Popular Alternatives “In the Wild”

* **ULID** (Universally Unique Lexicographically Sortable ID)

    * 128 bits; 48-bit time + 80-bit randomness.
    * 26 chars, Crockford Base32, lexicographically sortable.
* **KSUID** (K-Sortable Unique ID)

    * 160 bits; 32-bit timestamp + 128-bit randomness.
    * 27 chars, Base62.
* **NanoID**

    * Short, URL-safe IDs; not RFC UUIDs.
* **TypeID**

    * Based on UUIDv7 + human-readable type prefix.

---

## Go Libraries

* `github.com/google/uuid` — UUID v1–v5, v7 support, RFC 9562 compliant.
* `github.com/gofrs/uuid` — Actively maintained, supports v1–v8.
* `github.com/oklog/ulid/v2` — Reference ULID library.
* `github.com/segmentio/ksuid` — Reference KSUID implementation.
* `go.jetify.com/typeid/v2` — TypeID (UUIDv7 + prefix).
* `github.com/matoous/go-nanoid/v2` — NanoID implementation.

---

## Validation Options

### Regex (quick prefilter)

* **UUID (any v1–v8, RFC 4122/9562 variant):**

  ```regex
  (?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$
  ```

* **UUID (specific version v1–v5):**

  ```regex
  (?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$
  ```

* **UUID (specific version v6–v8):**

  ```regex
  (?i)^[0-9a-f]{8}-[0-9a-f]{4}-[678][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$
  ```

* **ULID:**

  ```regex
  ^[0-9A-HJKMNP-TV-Z]{26}$
  ```

* **KSUID:**

  ```regex
  ^[0-9A-Za-z]{27}$
  ```

* **NanoID (default 21 chars):**

  ```regex
  ^[A-Za-z0-9_-]{21}$
  ```

### Go Function (more correct)

Minimal validator without dependencies:

```go
func ValidateUUID(s string) (int, error) {
    if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
        return 0, errors.New("bad UUID shape")
    }
    hexStr := strings.ReplaceAll(s, "-", "")
    var b [16]byte
    if _, err := hex.Decode(b[:], []byte(hexStr)); err != nil {
        return 0, err
    }

    version := int(b[6] >> 4)
    variant := (b[8] & 0xE0) >> 5
    if variant != 0b100 {
        return 0, errors.New("invalid variant")
    }
    if version < 1 || version > 8 {
        return 0, errors.New("unknown version")
    }
    return version, nil
}
```

This checks:

* Shape (36-char with hyphens)
* Hex encoding
* Variant bits (must be RFC 4122/9562)
* Version 1–8

---

## Practical Guidance

* Use **regex** for lightweight validation in templates or configs.
* Use a **Go parser** if you need to enforce version/variant correctness.
* Use **libraries** if generating IDs or if supporting ULID/KSUID/TypeID.
* For modern apps: prefer **UUIDv7** (sortable, RFC-standard), or **ULID/KSUID** if you want lexicographically sortable, non-hyphenated strings.

---

**End of summary** — suitable for sharing with teammates when building `{<varname>:uuid:format[<uuid_format>]}` validators.
