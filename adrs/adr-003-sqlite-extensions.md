# ADR-003: Deterministic, Dependency-Aware Loading of SQLite Extensions

**Status:** Proposed

**Date:** 2025-09-02

**Summary:** Defines a deterministic, auditable mechanism to discover, select, download, verify, load, and initialize SQLite extensions — including **SQL-only** initialization hooks at DB level and per-extension level, dependency ordering, platform suffix resolution, operator controls (disable, pin, or autoload), and a strict **PRAGMA validation** (SQL linter) that enforces assignment form, phase gating (core-only vs extension-defined), and override policy. Hardcoded connection defaults are applied first, then DB `init_sql`, then per-extension init, then DB `post_load_sql`. macOS suffix reality is handled (`.so` preferred, `.dylib` fallback).

---

## 1. Goals

* **Determinism:** Same inputs ⇒ same outputs across machines/CI.
* **Safety:** Verified downloads (`sha256s`), explicit per-extension failure policy.
* **Operator control:** Disable by ID in config, pin versions, or `--no-autoload`.
* **Clarity:** Stable tie-breakers, thorough logging, explicit precedence.
* **Initialization:** SQL hooks plus strict PRAGMA linter for safety and consistency.
* **macOS reality:** Many SQLite extensions ship `.so` on Darwin; we try `.so` → `.dylib` when `${LIBEXT}` is used.

**Non-goals (for now):**

* SQL named-parameter binding (future scope if requested).
* Per-extension “requested open params” (future scope).
* OS/ARCH in filenames (these binaries are installed/run on a single platform).

---

## 2. Normative Keywords

The words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are to be interpreted as described in RFC 2119.

---

## 3. Terms & Defaults

* **Extension ID (`id`)**: `^[a-z0-9._-]+$` (lowercase recommended). Used in filenames, dependency references, disable directives.
* **Version (`version`)**: String; SemVer **recommended** (e.g., `v1.5.0`). Non-SemVer falls back to lexicographic ordering with a warning.
* **`LIBEXT` (library suffix)**: Linux → `.so`; **macOS → try `.so` then `.dylib`**; Windows → `.dll`.
* **Well-known directories** (searched in order):
  1. Project: `./.xmlui/sqlite3/exts/`
  2. User: `~/.config/xmlui/sqlite3/exts/`
* **Autoload**: Discovery from those directories without explicit config entries.
* **Hardcoded connection defaults** (applied immediately after open):
    ```
    PRAGMA journal_mode=WAL;
    PRAGMA busy_timeout=5000;
    PRAGMA synchronous=NORMAL;
    ```
  
---

## 4. Filenames & Manifests

### 4.1 Canonical binary filenames

* **Versioned (preferred):** `${ID}@${VERSION}${LIBEXT}`
  * Examples: `steampipe-github@v1.5.0.so` (Linux/macOS), `icu@73.2.dll` (Windows)
* **Unversioned (allowed, lower priority):**
  `${ID}${LIBEXT}`

> We **do not** include OS/ARCH in filenames. These binaries are installed for the current platform only.

### 4.2 Manifest Filenames

* Manifests **MUST** use the suffix: `*.sqlite3-ext.json`.
* There is no `*.sqlite-ext.json` support (to avoid fragmentation).

---

## 5. Variable Expansion

When expanding one of the `download_urls` as well as `binary_name`, `docs_url`, `repo_url` and/or `filepath`, the following variables are supported:

* `${OS}` ∈ `darwin | linux | windows` (lowercase)
* `${ARCH}` ∈ `amd64 | arm64` (lowercase)
* `${ID}` → extension id
* `${VERSION}` → manifest `version`
* `${LIBEXT}` → platform suffix with Darwin dual-try (`.so` then `.dylib`)

    * **linux:** `.so`
    * **darwin:** try `.so`, then `.dylib` (in this order)
    * **windows:** `.dll`

If a URL/path **omits** `${LIBEXT}` and hardcodes a suffix, that literal is used.

---

## 6. Database-Level Configuration (type=`sqlite3`)

```json
{
  "database": {
    "type": "sqlite3",

    "@note1": ["Appended to DSN as ?k=v&k2=v2 (e.g., { \"_foo\": \"1\" })"],
    "open_params": { "_allow_load_extension": "1" },

    "@note2": [
      "DB-level SQL runs once BEFORE any extensions load (init_sql)",
      "and once AFTER all extensions complete (post_load_sql)."
    ],
    "init_sql": [
      "/* DB SQL before any extensions load; PRAGMA assignments allowed here (core-only) */",
      "PRAGMA journal_mode=DELETE",
      "PRAGMA synchronous=FULL"
    ],
    "post_load_sql": [
      "/* DB SQL after all extensions have completed; PRAGMA assignments allowed (core & extension-defined) */",
       "PRAGMA trusted_schema=OFF",
       "PRAGMA cache_size=-200000"
    ],

    "@note3": ["See §7 for allowed entries."],
    "extensions": [ /* entries */ ]
  }
}
```

**Notes (normative):**

* `init_sql` executes **once** after hardcoded defaults and **before** any extension loads.
* `post_load_sql` executes **once** after **all** extensions finish their per-extension SQL.
* Any PRAGMAs in these SQL blocks are validated by the PRAGMA linter (see §12.3).

---

## 7. `extensions` Array — Entry Types

Each entry in `database.extensions` is exactly one of:

### 7.1 Disable by ID or Path

```json
{ "disable": "steampipe-plugin-github" }
```

Disables **all** instances of that ID (any version, manifest or autoload).

```json
{ "disable": "/path/to/specific/binary.so" }
```

Disables a **specific binary by path**. Path resolution supports:
- **Absolute**: `/absolute/path/to/binary.so`
- **Relative**: `./relative/path/binary.so` (relative to config file directory)  
- **Home-relative**: `~/sqlite3/exts/binary.so` (relative to user home)
- **Security**: Parent directory references (`..`) are **forbidden** for security

Disable applies **before** dependency resolution.

### 7.2 Inline manifest (full spec)

```json
{
  "id": "steampipe_sqlite_github",
  "title": "Steampipe GitHub",
  "version": "v1.5.0",
  "docs_url": "https://hub.steampipe.io/plugins/turbot/github",
  "repo_url": "https://github.com/turbot/steampipe-plugin-github",

  "download_urls": [
     "https://github.com/turbot/${ID}/releases/download/${VERSION}/${ID}.${OS}_${ARCH}.tar.gz",
     "https://cdn.example.com/sqlite3/exts/${ID}@${VERSION}${LIBEXT}"
  ],
  "sha256s": {
    "darwin_arm64": "sha256:…",
    "linux_amd64": "sha256:…",
    "windows_amd64": "sha256:…"
  },

  "entry_point": "sqlite3_extension_init",
  "depends_on": [],
  "load_order": 0,

  "@note1": ["on_failure can be one of error, warn, or ignore"],
  "on_failure": "warn",

  "vars": { "STEAMPIPE_CACHE": "false" },

  "pre_load_sql": [
    "/* runs BEFORE loading binary; PRAGMA assignments here must be core-only */",
    "ATTACH DATABASE ':memory:' AS extension_mem"
  ],

  "on_load_sql": [
    "// runs AFTER binary loaded; extension-defined PRAGMAs & module SQL allowed",
    "CREATE VIRTUAL TABLE IF NOT EXISTS gh_repos USING github_repos();"
  ]
}
```

### 7.3 Manifest Reference (Path or URL)
A Manifest Reference is only required when the extension is not found in a user's or project's well-known location, e.g. `<basePath>/sqlite3/exts`: 

```json
{ "manifest": "/path/to/your/exts/geojson.sqlite3-ext.json" }
```
Or:
```json
{ "manifest": "https://example.com/exts/geojson.sqlite3-ext.json" }
```

### 7.4 Shorthand include by ID

```json
{ "id": "fts5" }
```

* Meaning: “If present locally (manifest or binary), include it; do **not** download.”
* If **not found**, the loader **MUST WARN** (never silently ignore).

Optional constraint:

```json
{ "id": "fts5", "version": "v3.1.0" }
```

* Meaning: “Use that version if present; do not download; if absent, **WARN**.”

---

## 8. Manifest pseudo-schema (inline or referenced)

Any of `docs_url`, `repo_url`, `download_url[n]`, `filepath` may use ${OS}, ${ARCH}, ${ID}, ${VERSION}, ${LIBEXT}:

```json
{
  "id": "string, required, ^[a-z0-9._-]+$",
  "version": "string, required; SemVer recommended",
  "title": "string, optional descriptive name",

  "docs_url": "string, optional URL",
  "repo_url": "string, optional URL",

  "@note1": ["sha256 REQUIRED when download_url or download_urls is used"],
  "download_urls": [
    "string, optional"
  ],
  "filepath": "string, optional; if present and readable, used directly",

  "@note1": ["sha256 REQUIRED when download_url is used"],
  "sha256s": { 
    "darwin_arm64": "sha256:…", 
    "linux_amd64": "sha256:…", 
    "windows_amd64": "sha256:…" 
  },
  
  "entry_point": "string, default sqlite3_extension_init",
  "depends_on": ["id", "..."],
  "load_order": 0,

  "on_failure": "error | warn | ignore (default: warn)",

  "@note2": ["Optional env vars available to extension at load time (string values)"],
  "vars": { 
    "NAME": "VALUE", 
    "..." : "..." 
  },
  "@note3": [
    "vars are exported to the process environment",
    "if var_scope = app (default), they persist for the process lifetime",
    "if var_scope = load, set only during extension's load/init window"
  ],
  "var_scope": "string, one of load or app; default app"

  "@note3": ["Run prior to loading"],
  "pre_load_sql": ["SQL statement", "..."],

  "@note4": ["Run after loading"],
  "post_load_sql":  ["SQL statement", "..."]
}
```

**Integrity/Notes:**

- If `download_urls` is used, `sha256s["${OS}_${ARCH}"]` is **required** and must be lowercase keys like `darwin_arm64`.
  - Integrity MUST be verified post-download (and for preexisting files when loading).

- `filepath` MAY include `${LIBEXT}`; see macOS resolution behavior.
* **`vars`** are exported but not reset/unset unless `var_scope` set to `"load"`. In that case the process environment is set only for the brief window around load/initialization (set → load/init → unset), so extensions observing env at init can read them.

---

## 9. Discovery, Merging & Selection

### 9.1 Autoload scan

The loader **MUST** scan the well-known directories (in order) for:

* **Manifests:** `*.sqlite3-ext.json`
* **Ad-hoc binaries:** files matching
  `^([a-z0-9._-]+)(?:@([^/]+))?(\.so|\.dylib|\.dll)$`

Each discovered item yields a candidate:

* Manifests: as declared.
* Binaries without a manifest: synthesize a minimal manifest in memory:

    * `id` = filename base before `@`
    * `version` = parsed version or `"0.0.0-unspecified"`
    * `entry_point` = default
    * `on_failure` = `"warn"`
    * no pragmas/SQL/deps

### 9.2 Apply Disable Directives

Before any merging or dependency work, read `extensions` array, collect `{ "disable": "<id-or-path>" }` entries, and **remove** any candidates:
- If disable value matches an extension `id`: remove all instances of that ID (any version, manifest or autoload)
- If disable value is a path: resolve path (absolute, relative to config dir, or home-relative) and remove the specific binary at that resolved path
- **Security validation**: reject any path containing `..` parent references
If a disabled `id` is also explicitly declared as a manifest/manifest-ref in the same array, **disable wins** (log a clear note).

### 9.3 Precedence per ID

1. Inline manifest (authoritative)
2. Manifest reference (authoritative)
3. Shorthand by ID (local only; **no download**)
4. Autoload (local only)

### 9.4 Version Selection

**Pinned**
* If an authoritative manifest declares `version`, that version is **pinned**:
  * If `${ID}@${VERSION}${LIBEXT}` exists locally, verify hash (if `sha256s[os_arch]` present) and use it.
  * Else if `download_urls[n]` exists, **download → verify** (`sha256s[os_arch]`) → install as `${ID}@${VERSION}${LIBEXT}` → use.
  * Else apply `on_failure` (`error|warn|ignore`).

* **Unpinned**:

  * Prefer highest **SemVer** among local `${ID}@${VERSION}${LIBEXT}`; otherwise lexicographic with a warning.
  * Fallback to unversioned `${ID}${LIBEXT}`.
  * If shorthand `{ "id": … }` and nothing found, **WARN**.

### 9.5 Why a `latest` version token is **not** supported

We intentionally **do not** support `version: "latest"` (or `${VERSION}=latest`) for the following reasons:

1. **Integrity requires a version-bound hash.**
   Downloaded binaries **must** be verified against `sha256s["${OS}_${ARCH}"]`. Hashes are, by definition, tied to a **specific** artifact/version. A moving target like “latest” cannot be validated deterministically, which weakens our supply-chain guarantees.

2. **Reproducibility & audits.**
   Configs should produce the **same result** across time, machines, and CI. “Latest” introduces time-dependent behavior that can’t be reconstructed during incident response or audits.

3. **Deterministic dependency resolution.**
   Our topo-sorted init order and cross-extension SQL/PRAGMA phases assume stable capabilities. Auto-advancing one extension to “whatever is newest” can silently break dependents.

4. **Operator expectations & failure semantics.**
   When an extension changes, we want the change to be a **deliberate pin bump** (e.g., `id@v1.5.0 → id@v1.6.0`) so warnings/errors point to a concrete version, not a floating label.

5. **Attack surface reduction.**
   “Latest” is susceptible to redirection or repository compromise issues (TOCTOU). Requiring an explicit version + hash blocks this entire class of risks.

**Recommended practices instead**

* **Pin explicit versions** in manifests and let the loader fetch/verify that exact artifact (e.g., `steampipe-github@v1.5.0${LIBEXT}` + matching `sha256s`).
* If you truly want “use whatever is already on disk,” omit a manifest and rely on **autoload**; the loader will pick the **highest versioned file present** for that ID (or the unversioned file) from the well-known directories. This is **local and explicit**—it does not download—and the loader will **warn** when falling back to an unversioned file.
* Manage upgrades via CI or an administrative script that updates the pinned `version` and `sha256s` after you’ve reviewed release notes and validated in staging.

*If we ever consider a future “latest” mode, it would only be via a resolver that first discovers a concrete version **with hashes** (e.g., a signed “latest manifest” or versions index), then commits that resolved version—never by using the string “latest” at load time.*

---

## 10. macOS suffix resolution

When a path/URL contains **`${LIBEXT}`** on **darwin**:

1. Try `.so` (fetch/open/verify).
2. If that fails, try `.dylib`.

Literal suffixes are used as-is.

---

## 11. Dependencies & Ordering

* Build a directed graph: `id → depends_on`.
* Disabled nodes are removed.
* **Topologically sort** the remainder
* Tie-breakers: `load_order` (ascending), then `id` (lexicographic).

If a dependency is missing/disabled:

  * `on_failure:"error"` → **fail** with a clear reason.
  * `on_failure:"warn""` → **skip** that extension (and dependents), log warning.
  * `on_failure:"ignore"` → **skip** that extension (and dependents), no logging.

---

## 12. Initialization Order & PRAGMA Validation (SQL linter)


### 12.1 DB Connection-Level

On opening a SQLite connection, the loader applies hardcoded defaults (see §12.1.1), executes DB-level `init_sql`, then proceeds to extension discovery/initialization. After all extensions finish, the loader runs the DB-level post-load hook `post_load_sql`.

**Precedence:**

1. Hardcoded defaults
2. DB-level init\_sql
3. Per-extension (for each extension in topo order):
   3.1 pre\_load\_sql
   3.2 on\_load\_sql
4. DB-level post\_load\_sql

* **`open_params`** are appended to the DSN (e.g., `_allow_load_extension=1`).

### 12.1.1 Default Connection PRAGMAs (hardcoded; overridable)

On opening a SQLite connection, the loader applies the following **hardcoded defaults** before processing any `pre_sql` pragmas:

```
PRAGMA journal_mode=WAL;
PRAGMA busy_timeout=5000;
PRAGMA synchronous=NORMAL;
```

**Rationale:** WAL improves write concurrency; `busy_timeout=5000` reduces transient lock failures; `synchronous=NORMAL` is the recommended balance when WAL is enabled.

**Durability & logging:**
The loader reads back and logs the **effective** values returned by SQLite after applying the defaults (and any overrides), since `journal_mode` is persisted in the DB file and other settings may be connection-scoped. This makes drift visible in logs.

**Filesystem notes:**
1. On some network or unusual filesystems, WAL can be cranky. If someone ever runs on NFS/SMB and sees odd lock behavior, you may need a per-deployment override. Operators can override the default (e.g., revert to `journal_mode=DELETE`) when necessary.
2. `journal_mode` persists in the database file; others may not across connections. We’ll reapply them on open to guarantee the intended state.




### 12.2 PRAGMA validation (applies to all SQL blocks)

The loader inspects each SQL block and enforces PRAGMA safety:

1) **Detect PRAGMAs**
   Any statement matching `(?i)^\s*PRAGMA\s+` is treated as a PRAGMA. Split batches on `;` (respect quotes).

2) **Require assignments (no queries)**
   Accept only `PRAGMA <name>=<value>;` (flexible whitespace).
   **Reject** `PRAGMA <name>;` (query form) and any other form.

3) **Normalize for logging/conflicts**
   Compare `name` case-insensitively (`lower(name)`); keep `value` string as provided.

4) **Phase gating**
   - **DB `init_sql`**: **core SQLite** PRAGMAs only (no extension-defined PRAGMAs). Non-core here → **WARN** or **ERROR** (operator policy).
   - **Per-extension `pre_load_sql`**: **core-only** PRAGMAs (extension not loaded yet).
   - **Per-extension `on_load_sql`**: core and **extension-defined** PRAGMAs allowed.
   - **DB `post_load_sql`**: core and **extension-defined** PRAGMAs allowed.

5) **Override policy (`--allow-pragma-overrides`)**
   Maintain a map of first-set PRAGMA names. When a later statement sets the same name:
   - **Default (flag absent):**
     - If earlier was **DB-level**, **block** the change and **WARN** (`blocked_override id=<id> phase=<phase> name=<n> wanted=<X> kept=<Y>`).
     - If earlier was **extension-level**, allow it (extensions can override each other in topo order) and **WARN**.
   - **With flag present:** Allow the change (still **WARN**, `allowed_override old→new`).
   After DB `post_load_sql`, log **effective** final values for any PRAGMA names that changed.

6) **Value sanity checks (recommended)**
   Validate known domains (e.g., `journal_mode`/`synchronous` enums, integer ranges like `busy_timeout>=0`, booleanish {ON,OFF,1,0}). Unknown names are treated as extension-defined in `on_load_sql`/`post_load_sql`.

### 12.3 Transactions per block

- DB `init_sql`: one tx.
- Each extension’s `pre_load_sql`: one tx.
- Load binary (no tx).
- Each extension’s `on_load_sql`: one tx.
- DB `post_load_sql`: one tx.

Failures inside a block roll back that block and respect `on_failure`.

---

## 13. Loading & Environment

### 13.1 DB connection-level loading

For the selected order (from §11):

1. **Open DB** with `open_params`.
2. Apply **hardcoded defaults**.
3. Run **DB `init_sql`**, if present.
4. Proceed to the per-extension loop (below).
5. Run **DB `post_load_sql`**, if present.

At every step, apply `on_failure: "error" | "warn" | "ignore"` for that extension.

**Notes:**
* Pragmas and `init_sql` **SHOULD** execute inside **one transaction per extension** (default behavior).
* If `post_load_sql` depends on objects from another extension, that relationship **MUST** be expressed via `depends_on` so topo ordering enforces correctness.

### 13.2 Per-extension level loading

For each extension in topo order:

1. Run **`pre_load_sql`** (core-only PRAGMAs allowed here).
2. **Load the binary**:
    * If `filepath` provided, use that (may contain `${LIBEXT}`); if not present, fall back to download logic (if allowed by manifest).
    * If `download_urls` used, integrity **MUST** be verified via `sha256s[os_arch]`.
    * Use `entry_point` or default `sqlite3_extension_init`.
3. Run **`on_load_sql`** (core & extension-defined PRAGMAs, module SQL).
4. Apply `on_failure` policy on any error (`error` aborts; `warn` logs & continues; `ignore` continues silently). 

**Environment variables (`vars`)** (per extension):

- If `var_scope=="load"`, set vars immediately before step 1 and **restore/unset** them immediately after step 3.
- If `var_scope=="app"` (default), export vars to the process env for the process lifetime (required for lazily-reading extensions). Ideally extensions will use a unique prefix (e.g., `${ID}_…`) to minimize conflicts; log which mode is used.

---

## 14. Download, Verify & Install

* Build candidate URLs/paths using variables (§5).
* On darwin with `${LIBEXT}`, try `.so` then `.dylib`.
* For each candidate:
  * If remote: HEAD (if supported) then GET; save to temp.
  * Verify SHA256 against `sha256s["${OS}_${ARCH}"]`. On mismatch, discard and try the next candidate.
  * On match: install to the first writable well-known directory (prefer project-local) using filename `${ID}@${VERSION}${LIBEXT}`.
  * If **all** candidates fail to exist or verify: apply `on_failure`.

---

## 15. Security & Safety Considerations

- Defaults applied at open are: `journal_mode=WAL`, `busy_timeout=5000`, `synchronous=NORMAL` balance stability and developer ergonomics; ; operators can enforce PRAGMA policy in DB `init_sql` and `post_load_sql`.
* Consider hardening toggles such as PRAGMAs `foreign_keys=ON` and `trusted_schema=OFF`. These are **not** enabled by default to keep developer ergonomics high.
* Manifest MUST always provide downloaded URLs with hashes in `sha256s` and verify downloads via `sha256s`.
* Keep per-extension `on_failure` conservative (`warn`) for non-critical features, and `error` only when the extension is truly required.

---

## 16. Validation Rules

Loader MUST enforce:

- `id` matches `^[a-z0-9._-]+$`.
- `on_failure` ∈ {`error`,`warn`,`ignore`} (default `warn`).
- If `download_urls` present → require `sha256s[os_arch]`.
- `depends_on` must reference known (non-disabled) IDs; unresolved handled per §11.
- Multiple local versions: pick highest SemVer; else lexicographic with WARN.
- `load_order` is a tie-breaker only; must not override dependency ordering.
- Shorthand `{ "id": … }` not found → **WARN**.

_Note:_ PRAGMA validation occurs via SQL linter against the SQL blocks (see §12.2).

---

## 17. Operator Behavior & Logging

- Log each step at `info` (discover, select, download, verify, load, init).
- **Warn when shorthand `{ "id": "…" }` is not found** (never silent).
- For `warn` failures, include `id`, phase, and actionable reason.
- Provide a final “Loaded extensions (in order)” summary, including skipped/disabled with reasons.
- After applying defaults and DB `init_sql`, and again after DB `post_load_sql`, log **effective** `journal_mode`, `busy_timeout`, `synchronous`, plus any PRAGMA names that were overridden/blocked (showing final values) so operators can see if the database file or environment forced a different mode.

---

## 18. Examples

### 18.1 Disable one, pin one, autoload the rest

```json
"extensions": [
  { "disable": "legacy-vtable" },

  {
    "id": "steampipe-github",
    "version": "v1.5.0",
    "download_url": "https://cdn.example.com/${ID}/${VERSION}/${ID}@${VERSION}${LIBEXT}",
    "sha256s": { "darwin_arm64": "sha256:…", "linux_amd64": "sha256:…" },
    "on_failure": "warn",
    "pre_load_sql": [
      "PRAGMA trusted_schema=OFF;"
    ],
    "on_load_sql": [
      "CREATE VIRTUAL TABLE IF NOT EXISTS gh_repos USING github_repos();"
    ]
  },
  "@note1":["No explicit entry for 'csv' — if csv@*.so exists in the dirs, it autoloads."]
]
```

### 18.2 Shorthand include with warn-if-missing, plus a manifest reference

Warn if fts5 not found:

```json
"extensions": [
  { "id": "fts5" },
  { "manifest": "./exts/geojson.sqlite3-ext.json" }
]
```

### 18.3 DB-level SQL before/after all extensions

```json
{
  "database": {
    "type": "sqlite3",
    "open_params": { "_allow_load_extension": "1" },

    "init_sql": [
      "PRAGMA journal_mode=WAL;",
      "PRAGMA busy_timeout=5000;",
      "PRAGMA synchronous=NORMAL;"
    ],

    "post_load_sql": [
      "PRAGMA trusted_schema=OFF;"
    ],

    "extensions": [
      { "id": "fts5" },
      { "disable": "legacy-vtable" }
    ]
  }
}
```

### 18.4 Inline with deps and on-load SQL

```json
"extensions": [
  {
    "id": "mytokenizer",
    "version": "1.2.0",
    "download_url": "https://dl.example.com/${ID}/${ID}@${VERSION}${LIBEXT}",
    "sha256s": {
      "darwin_arm64": "sha256:…",
      "linux_arm64": "sha256:…"
    },
    "depends_on": ["icu"],
    "on_failure": "error",
    "pre_load_sql": [
      "PRAGMA trusted_schema=OFF",
      "PRAGMA case_sensitive_like=ON"
    ],
    "on_load_sql": [
      "CREATE VIRTUAL TABLE IF NOT EXISTS toks USING mytokenizer_vtab();"
    ]
  }
]
```

---

## 19. Alternatives Considered

- **Separate PRAGMA arrays and phases:** Replaced with SQL-only blocks to avoid proliferation and to align with how users copy/paste from docs. The PRAGMA linter preserves safety guarantees without schema complexity.
- **OS/ARCH in filenames:** Rejected; binaries are installed on a single platform.
- **“latest” pseudo-version:** Rejected; see §9.5. 

---

## 20. Consequences

- Deterministic, reproducible, and auditable initialization.
- Clear operator control (disable, pin, autoload) and visibility (WARN on missing shorthand).
- Safe defaults with explicit, centralized SQL hooks.
- Strong PRAGMA validation with override policy and final state logging.
- Operator-friendly logs, including **WARN if shorthand include not found** (never silent).

---

## 22. Open Questions / Future Scope

- SQL parameter binding for SQL.
- musl/glibc specificity in `sha256s` keys (e.g., `linux_amd64_musl`).

---
