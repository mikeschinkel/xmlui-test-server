# UUID / GUID Validation & Libraries (Go-centric)

This document summarizes algorithms, UUID versions, and Go libraries relevant to implementing a validator for path/query variables of the form:

```
{<varname>:uuid:format[<uuid_format>]}
```

where `<uuid_format>` might be `v4`, `v7`, `ulid`, `ksuid`, etc.

---

## UUID Versions (RFC 4122 / RFC 9562)

* **`v1`** – Time + MAC address (sortable but leaks MAC).
* **`v2`** – Time + POSIX UID/GID (rare).
* **`v3`** – Name-based, MD5.
* **`v4`** – Random (most common, Microsoft GUID style).
* **`v5`** – Name-based, SHA-1.
* **`v6`** – Reordered `v1` (time-ordered, DB-friendly).
* **`v7`** – Unix timestamp (ms) + random (DB-friendly, modern default).
* **`v8`** – Custom / experimental fields.

RFC 9562 (2024) standardized **`v6`–`v8`**.

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

* `github.com/google/uuid` — UUID `v1`–`v5`, `v7` support, RFC 9562 compliant.
* `github.com/gofrs/uuid` — Actively maintained, supports `v1`–`v8`.
* `github.com/oklog/ulid/`v2`` — Reference ULID library.
* `github.com/segmentio/ksuid` — Reference KSUID implementation.
* `go.jetify.com/typeid/`v2`` — TypeID (UUIDv7 + prefix).
* `github.com/matoous/go-nanoid/`v2`` — NanoID implementation.

---

## Validation Options

### Regex (quick prefilter)

* **UUID (any `v1`–`v8`, RFC 4122/9562 variant):**

  ```regex
  (?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$
  ```

* **UUID (specific version `v1`–`v5`):**

  ```regex
  (?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$
  ```

* **UUID (specific version `v6`–`v8`):**

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

## Examples


### UUID Generic
Represents a canonical, version-agnostic UUID example.

Example: `01234567-890A-BCDE-F012-34567890ABCD`

This example simply increments each hexadecimal digit in sequence, starting from `0` through `F`, then repeating.

---

### UUIDv1 _(Time-based with MAC address)_

Uses a timestamp and the host machine’s MAC address, making it potentially non-anonymous.  

Example: `f81d4fae-7dec-11d0-a765-00a0c91e6bf6`

This is the canonical example from the original UUID specification (RFC 4122). Its embedded timestamp corresponds to approximately **1994-11-04 08:48:37 UTC**, illustrating how early UUIDv1s recorded precise creation times down to sub-microsecond resolution.  
Because the lower bits also include the node’s MAC address (`00:a0:c9:1e:6b:f6`), such UUIDs can reveal both *when* and *where* they were created.

The timestamp encodes the number of 100-nanosecond intervals since the Gregorian epoch (`1582-10-15 00:00:00 UTC`), combined with clock sequence and MAC address bits to ensure uniqueness.



---

### UUIDv2 _(DCE Security)_
Replaces part of the time field with POSIX UIDs or GIDs for distributed-computing environments. It is rarely used and not fully specified in any current RFC. There is no widely known canonical example.

Example: `02010203-0405-2607-8809-0a0b0c0d0e0f`

This is a handcrafted, representative example that follows the UUIDv2 format (version digit `2`). Its obscurity illustrates how uncommon this variant is in practice.

---

### UUIDv3 _(Name-based, `MD5` hash)_
Derives the UUID deterministically from a namespace UUID and a name string using an `MD5` hash. The DNS namespace is well-known.

Example: `b1e42d72-b5b4-3286-9a25-e78457c1543b`

This UUID is deterministically generated from the standard DNS namespace and the hostname `www.example.com`. Any developer can reproduce this exact value.

---

### UUIDv4 _(Random)_
Generated from random bits, with only a few bits reserved for version and variant. Any UUID with a `4` in the third group is a `v4`.

Example: `deadbeef-cafe-4011-8123-b1d5c0d51234`

This example uses hexadecimal _“leet speak”_ for recognizability; UUIDv4s are intentionally unpredictable, so memorable patterns are fine for illustration.

---

### UUIDv5 _(Name-based, `SHA-1` hash)_
Similar to `v3` but uses `SHA-1` instead of `MD5` for hashing.

Example: `2a98f1f0-0a71-50e5-9d51-8650e68d9518`

This UUID is deterministically generated from the standard URL namespace and the URL `https://www.example.com`.

---

### UUIDv6 _(Reordered Time-based)_

Reorders the fields of a UUIDv1 so that the timestamp occupies the most significant bits, making the resulting UUIDs lexicographically sortable while preserving backward compatibility with `v1` semantics.

Example: `1ec7c816-e5f6-6b2d-98ac-b0b3d64c1533`

A syntactically correct, handcrafted placeholder showing version digit `6`.  

If generated according to the `v6` layout, the leading timestamp portion `1ec7c816e5f6` would correspond roughly to `2024-03-11 15:42:00 UTC` depending on the exact clock and node values used.  

As with `v1`, `v6` embeds both timestamp and node information, but its reordered format enables efficient chronological sorting in databases or logs.

---
### UUIDv7 _(Unix-epoch Time-based)_

Introduced to align UUIDs with the Unix epoch time format while maintaining lexicographic sort order.  

Example: `018d9f10-5341-7c91-9e73-b3c14d9b4b0e`

This example follows a typical `v7` implementation.  
The leading timestamp portion (`018d9f10_5341`, in hexadecimal) corresponds to approximately **2025-01-29 14:00:00 UTC**, meaning this UUID could have been generated at that instant.  

UUIDv7 encodes the number of milliseconds since `1970-01-01 00:00:00 UTC` in its most significant 48 bits, followed by 74 bits of randomness. This design ensures that newer UUIDs sort after older ones while remaining globally unique. Because UUIDv7s are time-based, their timestamps can be resolved to the exact millisecond of creation while still including a large random component to avoid collisions across systems.

---

### UUIDv8 _(Custom)_
Reserved for vendor- or application-defined structures. The version digit `8` indicates a custom format.

Example: `20251018-b26a-8025-a12b-4c5d6e7f8a9b`

Here the leading eight hexadecimal digits (`20251018`) encode a generation date _(Oct 18 2025),_ followed by random payload data—illustrating `v8`’s flexibility.

---

### Combined Types
Some systems group UUIDs by version ranges.

#### UUID Generic
As noted earlier, the _"nil"_ UUID (`00000000-0000-0000-0000-000000000000`) serves as a universal placeholder.

#### UUID `v1`–`v5`
You can reference the canonical `v1` example: `f81d4fae-7dec-11d0-a765-00a0c91e6bf6`.

#### UUID `v6`–`v8`
No standard examples exist yet; use a `v6` placeholder such as the one above.



### KSUID (K-Sortable Unique ID)

KSUID is a time-sortable, globally unique identifier developed by a private company to meet internal needs and later released as open source.

Example: `0ujsswThIGTUYm2K8FjOOfXtY1K`

This KSUID is the first example in the [`segmentio/ksuid`](https://github.com/segmentio/ksuid) repository’s `README.md`.  
Its embedded timestamp represents `2017-10-09 21:00:00 UTC`, which is `123,475,966` seconds after KSUID’s custom epoch of `2014-05-13 00:00:00 UTC`.

KSUID combines a timestamp with a random payload, producing lexicographically sortable identifiers without requiring a central authority. KSUID is widely used in distributed systems but is not an official standard.

> **Note:** A KSUID is `20` bytes in binary: a `4`-byte big-endian timestamp followed by `16` bytes of cryptographically random data.  
> When encoded as a `27`-character Base62 string, the entire 20-byte value is treated as a single large integer.  
> Therefore, although the timestamp occupies the first 4 bytes in binary form, it does **not** correspond to a fixed number of leading characters in the Base62 string.

### NanoID

NanoID is a compact, secure, URL-safe identifier generator designed as a modern alternative to UUIDs. It prioritizes brevity, performance, and strong randomness while remaining collision-resistant at scale. Unlike KSUID, NanoID does not include a timestamp component—each ID is purely random and therefore not time-sortable.

Example: `V1StGXR8_Z5jdHi6B-myT`

This example is from the first example in `the README.md` at the [ai/nanoid](https://github.com/ai/nanoid) GitHub repository. This `21`-character NanoID uses the default `64`-character alphabet — `A–Z`, `a–z`, `0–9`, `_`, `-` — providing approximately `128` bits of entropy—comparable to `UUID `v4`` but in a shorter, URL-friendly form. Developers can also define custom alphabets and lengths, allowing generation of IDs that are compact, human-readable, or thematically styled for docs or testing.

NanoID originated as an open-source project to provide a small, dependency-free library suitable for browsers, Node.js, and other environments. It has since become a de-facto standard for short, human-safe identifiers.

> **Note:** Because NanoIDs are random rather than time-based, they cannot be sorted chronologically.

### ULID _(Universally Unique Lexicographically Sortable Identifier)_

A ULID is a `128`-bit, lexicographically sortable identifier that combines a `48`-bit millisecond Unix-epoch timestamp with `80` bits of randomness.  

Example: `01G65Z755AFWAKHE12NY0CQ9FH`

This ULID appears as the first example in the [`ulid/spec`](https://github.com/ulid/spec) GitHub repository. The timestamp portion (`01G65Z755A`) corresponds to approximately `2022-07-09 01:52:21 UTC`, encoding the number of milliseconds since the Unix epoch. The remaining characters contain 80 bits of cryptographically secure randomness, ensuring uniqueness even when multiple ULIDs are generated in the same millisecond.

> **Note:** ULIDs use Crockford’s Base32 alphabet — `0–9`, `A–Z`, excluding `I`, `L`, `O`, and `U` — for human readability.  
> Their design influenced later formats such as UUID `v7`, which similarly embeds sortable time information within a 128-bit structure.

### CUID (Collision-resistant Unique Identifier)

A CUID is a human-readable, lexicographically sortable identifier designed to reduce the likelihood of collisions in distributed systems.  
Example: `ckf0f9e5x0000q3yz4dq7a1qf`

This example comes from the [`paralleldrive/cuid`](https://github.com/paralleldrive/cuid) repository. The leading character (`c`) denotes the format type, and the next sequence (`kf0f9e5x0`) encodes a timestamp derived from the Unix epoch, representing approximately **2023-06-29 18:45:00 UTC**.  
The remaining characters include a host fingerprint, process counter, and random bits, which together ensure uniqueness even across concurrent instances and multiple machines.

CUID was originally developed as an open-source project to improve upon UUIDs for database use, ensuring strong uniqueness while remaining short, URL-safe, and monotonic within a single process.  
CUIDs are commonly used in web applications and databases where predictable, sortable IDs are preferred over purely random ones.

> **Note:** CUIDs are 25 characters long and consist of lowercase letters and digits, making them URL- and database-friendly.  
> Because the timestamp prefix is monotonic, CUIDs are sortable by creation time, similar to ULID and KSUID, but they emphasize simplicity and collision resistance over strict cryptographic entropy.

### Snowflake ID _(Time-ordered 64-bit Identifier)_

A **Snowflake ID** is a`64`-bit, time-ordered unique identifier originally developed by Twitter to generate sortable IDs at massive scale without central coordination.  Each Snowflake encodes a timestamp, a machine identifier, and a per-process sequence number, enabling decentralized ID generation that remains chronologically ordered.


Example: `1888944671579078978`

This example appears in [Wikipedia’s article on Snowflake IDs](https://en.wikipedia.org/wiki/Snowflake_ID) and corresponds to a tweet published by **@Wikipedia** on `February 10, 2025, at 13:34:39.256 UTC`. 

The first 41 bits represent the number of milliseconds since **Twitter’s custom epoch**, which began at `2010-11-04 01:42:54.657 UTC` — Unix time `1288834974657`. Adding the timestamp value `450359504599` to that epoch yields the tweet’s precise creation time.  

The next 10 bits identify the machine (or datacenter), and the final 12 bits encode the sequence number (`322`), indicating that this was the 322nd Snowflake generated during that millisecond.

> **Note:** Because the timestamp occupies the most significant bits, Snowflake IDs naturally sort in chronological order when compared numerically.  
> Many distributed systems—including **Discord**, **Instagram**, and various open-source projects—have adopted the Snowflake pattern to generate scalable, time-based identifiers suitable for databases, message queues, and event logs.

---

## Practical Guidance

* Use **regex** for lightweight validation in templates or configs.
* Use a **Go parser** if you need to enforce version/variant correctness.
* Use **libraries** if generating IDs or if supporting ULID/KSUID/TypeID.
* For modern apps: prefer **UUIDv7** (sortable, RFC-standard), or **ULID/KSUID** if you want lexicographically sortable, non-hyphenated strings.

---

**End of summary** — suitable for sharing with teammates when building `{<varname>:uuid:format[<uuid_format>]}` validators.
