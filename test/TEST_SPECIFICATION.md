# Integration Test Specification

**Status:** Active Development
**Last Updated:** 2025-10-11
**Total Test Scenarios:** 34

---

## 1. Test Status Overview

### Summary by Status

| Status | Count | Tests |
|--------|-------|-------|
| ✅ **PASSING** | 11 | 02, 03, 04, 06, 07, 08, 10, 11, 14, 18, 26 |
| ⚠️ **PARTIAL** | 1 | 09 (validation not enforced) |
| ❌ **FAILING** | 6 | 05, 20, 21, 22, 29, 30 |
| 💥 **CRASHING** | 6 | 12, 13, 15, 16, 17, 31, 34 |
| ❓ **UNTESTED** | 10 | 01, 19, 23, 24, 25, 27, 28, 32, 33 |

### Priority Matrix

| Priority | Issue | Affected Tests | Action Required |
|----------|-------|----------------|-----------------|
| 🔥 **CRITICAL** | Parameter type validation not implemented | All negative validation tests | Implement type validators for int, uuid, slug, date, alphanumeric |
| 🔥 **CRITICAL** | Query parameter extraction missing | 20-22, 31, 34 | Implement query string parser |
| 🔥 **CRITICAL** | Constraint naming crash (hyphen in `not-empty`) | 16, 17, 31, 34 | ✅ FIXED: Renamed to `notempty` |
| 🚨 **HIGH** | Enum constraint crashes URL parser | 12 | Fix comma handling in enum values |
| 🚨 **HIGH** | Regex constraint crashes URL parser | 13 | Fix special character escaping |
| 🚨 **HIGH** | UUID format constraint not recognized | 15 | Implement `format[v4]` constraint |
| ⚠️ **MEDIUM** | Boolean conversion not implemented | 05 | Convert "true"/"false" → 1/0 for SQLite |
| ⚠️ **MEDIUM** | Constraint validation not enforced | 09-17 | Implement constraint checking |

---

## 2. Test Categories

### Category 1: Data Types (Tests 01-08)

Basic parameter type validation without constraints.

#### Test 01: Integer Type (`{id:int}`)
- **Status:** ❓ UNTESTED (commented out)
- **Path:** `GET /users/{id:int}`
- **Valid:** `/api/users/1` → 200, returns Alice Carter
- **Invalid:** `/api/users/abc` → 422, RFC 9457 Invalid Parameter Type
- **Invalid:** `/api/users/12.34` → 422, RFC 9457 Invalid Parameter Type
- **Blocker:** Type validation not implemented

#### Test 02: String Type (`{email:string}`)
- **Status:** ✅ PASSING
- **Path:** `GET /users/by-email/{email:string}`
- **Valid:** `/api/users/by-email/alice@example.com` → 200

#### Test 03: UUID Type (`{uuid:uuid}`)
- **Status:** ✅ PASSING (valid cases only)
- **Path:** `GET /users/by-uuid/{uuid:uuid}`
- **Valid:** `/api/users/by-uuid/550e8400-e29b-41d4-a716-446655440001` → 200
- **Invalid:** `/api/users/by-uuid/not-a-valid-uuid` → 400 (should be 422)
- **Invalid:** `/api/users/by-uuid/550e8400e29b41d4a716446655440001` → 400 (missing hyphens)
- **Blocker:** Format validation not implemented

#### Test 04: Slug Type (`{slug:slug}`)
- **Status:** ✅ PASSING (valid cases only)
- **Path:** `GET /users/by-slug/{slug:slug}`
- **Valid:** `/api/users/by-slug/alice-carter` → 200
- **Invalid:** `/api/users/by-slug/alice carter` → 400 (spaces not allowed)
- **Invalid:** `/api/users/by-slug/alice@carter` → 400 (special chars not allowed)
- **Blocker:** Format validation not implemented

#### Test 05: Boolean Type (`{active:boolean}`)
- **Status:** ❌ FAILING
- **Path:** `GET /users/by-status/{active:boolean}`
- **Issue:** Returns 404 "No results found" when querying with "true"/"false"
- **Root Cause:** Boolean URL parameters need conversion to SQLite INTEGER (1/0)
- **Action:** Implement boolean type converter

#### Test 06: Real/Decimal Type (`{rating:real}`)
- **Status:** ✅ PASSING (valid cases only)
- **Path:** `GET /users/by-rating/{rating:real}`
- **Valid:** `/api/users/by-rating/4.0` → 200
- **Invalid:** `/api/users/by-rating/abc` → 400 (should be 422)
- **Blocker:** Type validation not implemented

#### Test 07: Date Type (`{birth_date:date}`)
- **Status:** ✅ PASSING (valid cases only)
- **Path:** `GET /users/by-birth-date/{birth_date:date}`
- **Valid:** `/api/users/by-birth-date/1990-05-15` → 200
- **Invalid:** `/api/users/by-birth-date/05-15-1990` → 400 (wrong format)
- **Invalid:** `/api/users/by-birth-date/not-a-date` → 400 (invalid value)
- **Blocker:** Format validation not implemented

#### Test 08: Alphanumeric Type (`{sensor_id:alphanumeric}`)
- **Status:** ✅ PASSING (valid cases only)
- **Path:** `GET /measurements/by-sensor/{sensor_id:alphanumeric}`
- **Valid:** `/api/measurements/by-sensor/TEMP001` → 200
- **Invalid:** `/api/measurements/by-sensor/TEMP-001` → 400 (hyphen not allowed)
- **Invalid:** `/api/measurements/by-sensor/TEMP 001` → 400 (space not allowed)
- **Blocker:** Format validation not implemented

---

### Category 2: Constraints (Tests 09-17)

Parameter constraints and validation rules.

#### Test 09: Integer Range Constraint (`{score:int:range[0..100]}`)
- **Status:** ⚠️ PARTIAL (accepts invalid values)
- **Path:** `GET /users/by-score/{score:int:range[0..100]}`
- **Valid:** `/api/users/by-score/85` → 200
- **Valid:** `/api/users/by-score/50` → 200 (no results but valid)
- **Invalid:** `/api/users/by-score/150` → 400 (should be 422, currently returns 404)
- **Invalid:** `/api/users/by-score/-10` → 400 (should be 422, currently returns 404)
- **Issue:** Constraint validation not enforced
- **Action:** Implement range constraint validator

#### Test 10: Real Range Constraint (`{rating:real:range[0.0..5.0]}`)
- **Status:** ✅ PASSING (valid cases only)
- **Path:** `GET /users/by-rating-range/{rating:real:range[0.0..5.0]}`
- **Valid:** `/api/users/by-rating-range/4.5` → 200
- **Note:** Constraint validation not tested yet

#### Test 11: Length Constraint (`{slug:slug:length[5..50]}`)
- **Status:** ✅ PASSING (valid cases only)
- **Path:** `GET /users/by-slug-length/{slug:slug:length[5..50]}`
- **Valid:** `/api/users/by-slug-length/alice-carter` → 200
- **Note:** Constraint validation not tested yet

#### Test 12: Enum Constraint (`{status:string:enum[active,archived,draft]}`)
- **Status:** 💥 CRASHING
- **Path:** `GET /projects/by-status/{status:string:enum[active,archived,draft]}`
- **Error:** "invalid URL path" - commas in enum values break URL parser
- **Action:** Fix comma handling in constraint parser

#### Test 13: Regex Constraint (`{email:string:regex[^[^@]+@[^@]+\\.[^@]+$]}`)
- **Status:** 💥 CRASHING
- **Path:** `GET /users/by-email-pattern/{email:string:regex[...]}`
- **Error:** "invalid URL path" - regex special characters break URL parser
- **Action:** Fix regex pattern escaping in parser

#### Test 14: Date Format Constraint (`{birth_date:date:format[yyyy-mm-dd]}`)
- **Status:** ✅ PASSING
- **Path:** `GET /users/by-date/{birth_date:date:format[yyyy-mm-dd]}`
- **Valid:** `/api/users/by-date/1990-05-15` → 200

#### Test 15: UUID Format Constraint (`{uuid:uuid:format[v4]}`)
- **Status:** 💥 CRASHING
- **Path:** `GET /users/by-uuid-v4/{uuid:uuid:format[v4]}`
- **Error:** "unknown constraint type: constraint_spec=v4"
- **Issue:** The `format[v4]` constraint type is not recognized
- **Action:** Implement UUID format constraint

#### Test 16: NotEmpty Constraint (`{email:string:notempty}`)
- **Status:** ✅ FIXED (was 💥 CRASHING)
- **Path:** `GET /users/search-email?{email:string:notempty}`
- **Previous Issue:** Hyphen in "not-empty" crashed parser
- **Fix:** Renamed to `notempty` (all lowercase, no special chars)
- **Note:** Query parameter extraction still not implemented

#### Test 17: Multiple Constraints (`{name:string:notempty,length[3..50]}`)
- **Status:** ✅ FIXED (was 💥 PANICKING)
- **Path:** `GET /users/by-name/{name:string:notempty,length[3..50]}`
- **Previous Issue:** Contained "not-empty" constraint with hyphen
- **Fix:** Renamed to `notempty`
- **Note:** Constraint validation still not enforced

---

### Category 3: Implicit Types & Query Parameters (Tests 18-22)

Type inference and query string parameter extraction.

#### Test 18: Implicit Type Inference (`{int}`)
- **Status:** ✅ PASSING
- **Path:** `GET /implicit/{int}`
- **Valid:** `/api/implicit/1` → 200
- **Note:** Implicit type inference for "int" works correctly

#### Test 19: Implicit Type with Double Colon (`{decimal::range[0.0..1000.0]}`)
- **Status:** ❓ UNTESTED
- **Path:** `GET /implicit-decimal/{decimal::range[0.0..1000.0]}`
- **Note:** Needs dedicated test run

#### Test 20: Basic Query Parameter (`?{q:string}`)
- **Status:** ❌ FAILING
- **Path:** `GET /users/search?{q:string}`
- **URL:** `/api/users/search?q=alice`
- **Error:** "Database parameters not found: check the logs for details (Status: 400)"
- **Issue:** Query parameters are NOT extracted from URL query string
- **Root Cause:** System tries to extract from JSON body instead
- **Action:** Implement query parameter extraction

#### Test 21: Optional Query Parameters with Defaults (`?{limit?10:int}&{offset?0:int}`)
- **Status:** ❌ LIKELY FAILING
- **Path:** `GET /users/paginate?{limit?10:int}&{offset?0:int}`
- **Issue:** Same as test 20 (query extraction not implemented)

#### Test 22: Path + Query Combined (`GET /projects/{project_id:int}/tasks?{status:string}`)
- **Status:** ❌ LIKELY FAILING
- **Path:** `GET /projects/{project_id:int}/tasks?{status:string}`
- **Issue:** Same as test 20 (query extraction not implemented)

---

### Category 4: Advanced Features (Tests 23-25)

Multi-segment parameters and JSON body extraction.

#### Test 23: Multi-Segment Parameter (`{file_path*}`)
- **Status:** ❓ UNTESTED
- **Path:** `GET /logs/by-path/{file_path*}`
- **Note:** Multi-segment capture with `*` suffix

#### Test 24: POST with JSON Body Parameters
- **Status:** ❓ UNTESTED
- **Method:** `POST /users/{id:int}/update`
- **Body:** `{"name": "Alice Updated", "email": "alice.updated@example.com"}`
- **Note:** JSON body parameter extraction

#### Test 25: PUT with Nested JSON Body (`{task.title}`)
- **Status:** ❓ UNTESTED
- **Method:** `PUT /projects/{project_id:int}/tasks`
- **Body:** `{"task": {"title": "New Task", "details": "...", ...}}`
- **Note:** Nested JSON parameter extraction

---

### Category 5: Response Handling (Tests 26-28)

Cardinality and row type variations.

#### Test 26: Cardinality One
- **Status:** ✅ PASSING
- **Path:** `GET /users/{id:int}`
- **Config:** `"cardinality": "one"`
- **Note:** Response format (object vs array) not verified in test

#### Test 27: Cardinality Many
- **Status:** ✅ LIKELY PASSING
- **Path:** `GET /users/active`
- **Config:** `"cardinality": "many"`
- **Note:** Response format (array) not verified in test

#### Test 28: Row Type String
- **Status:** ❓ UNTESTED
- **Path:** `GET /hello`
- **Config:** `"row_type": "string"`
- **Note:** Plain text response format

---

### Category 6: Error Handling (Tests 29-33)

HTTP error responses and RFC 9457 format.

#### Test 29: Invalid Parameter Type Error
- **Status:** ❌ FAILING
- **Path:** `GET /users/abc` (expecting int)
- **Expected:** 422 status with RFC 9457 error
- **Actual:** Returns 404, not 422
- **Issue:** Type validation not implemented

#### Test 30: Constraint Violation Errors
- **Status:** ❌ FAILING
- **Path:** `GET /users/by-score/150` (range is [0..100])
- **Expected:** 422 status with RFC 9457 error
- **Actual:** Returns 404, not 422
- **Issue:** Constraint validation not implemented

#### Test 31: Missing Required Parameter Error
- **Status:** 💥 PANICS (was due to hyphen, now likely fails differently)
- **Path:** `GET /users/search?{email:string:notempty}` (no query param provided)
- **Expected:** 422 status with RFC 9457 error
- **Issue:** Query parameter extraction not implemented

#### Test 32: 404 Not Found
- **Status:** ❓ UNTESTED
- **Path:** `GET /api/nonexistent`
- **Expected:** 404 status
- **Note:** Should return 404 for unknown endpoints

#### Test 33: 405 Method Not Allowed
- **Status:** ❓ UNTESTED
- **Path:** `POST /api/users/1` (GET-only endpoint)
- **Expected:** 405 status
- **Note:** Should return 405 for wrong HTTP method

---

### Category 7: Complex Scenarios (Test 34)

#### Test 34: Multiple Features Combined
- **Status:** 💥 PANICS (due to multiple blockers)
- **Path:** `GET /tasks/search/{project_id:int:range[1..100]}?{q:string:notempty}&{limit?10:int:range[1..100]}&{offset?0:int}`
- **URL:** `/api/tasks/search/1?q=design&limit=5`
- **Blockers:**
  - Query parameter extraction not implemented
  - Constraint validation not enforced
  - ~~"not-empty" constraint causes panic~~ ✅ FIXED

---

## 3. RFC 9457 Error Response Reference

### Error Type URLs

All error type URLs use the `schema.xmlui.org` domain:

```
https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type
https://schema.xmlui.org/errors/test-server/api/validation/constraint-violation
https://schema.xmlui.org/errors/test-server/api/validation/missing-required-parameter
https://schema.xmlui.org/errors/test-server/api/validation/invalid-format
https://schema.xmlui.org/errors/test-server/api/database/query-failed
https://schema.xmlui.org/errors/test-server/api/database/no-results
https://schema.xmlui.org/errors/test-server/api/endpoint-not-matched
https://schema.xmlui.org/errors/test-server/api/method-not-allowed
```

### Status Code Usage

| Status Code | Use Case | Example |
|-------------|----------|---------|
| **422** | Type validation failure | `"abc"` for `{id:int}` parameter |
| **422** | Constraint violation | `150` for `{score:int:range[0..100]}` |
| **422** | Format validation failure | `"not-a-uuid"` for `{uuid:uuid}` |
| **422** | Missing required parameter | Query parameter marked required but absent |
| 400 | Malformed request syntax | Invalid JSON body, malformed headers |
| 404 | Resource not found | Valid request but resource doesn't exist in database |
| 405 | Method not allowed | POST to GET-only endpoint |
| 500 | Internal server error | Unexpected errors, database failures |

### Error Response Examples

All error responses follow RFC 9457 with standard fields (`type`, `title`, `status`, `detail`, `instance`) and optional `extensions` array for custom validation details.

**Standard RFC 9457 Fields:**
- `type` - Error type URI
- `title` - Short human-readable summary
- `status` - HTTP status code
- `detail` - Human-readable explanation
- `instance` - URI reference identifying the specific occurrence

**Custom Extensions** (in `extensions` array):
- `parameter` - Parameter name that caused the error
- `expected_type` - Expected data type (e.g., "integer", "uuid")
- `received_value` - Actual value received
- `location` - Where parameter appears ("path", "query", "body")
- `constraint` - Constraint specification (e.g., "range[0..100]")
- `suggestion` - Helpful suggestion for fixing the error
- `validation_errors` - Array of validation errors for complex cases

#### 1. Invalid Parameter Type (Status: 422)

**Test:** `get_user_by_int_id_invalid_string`
**Path:** `/api/users/abc`

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type",
  "title": "Invalid Parameter Type",
  "status": 422,
  "detail": "Parameter 'id' expected an integer type but got 'abc'",
  "instance": "/api/users/abc",
  "extensions": [
    {
      "parameter": "id",
      "expected_type": "integer",
      "received_value": "abc",
      "location": "path",
      "suggestion": "Use an integer for 'id' like 123, for example: /api/users/123"
    }
  ]
}
```

#### 2. Constraint Violation (Status: 422)

**Test:** `get_users_by_score_invalid_too_high`
**Path:** `/api/users/by-score/150`

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/validation/constraint-violation",
  "title": "Constraint Violation",
  "status": 422,
  "detail": "Parameter 'score' value 150 violates constraint range[0..100]",
  "instance": "/api/users/by-score/150",
  "extensions": [
    {
      "parameter": "score",
      "received_value": "150",
      "location": "path",
      "constraint": "range[0..100]"
    }
  ]
}
```

#### 3. Invalid Format (Status: 422)

**Test:** `get_user_by_uuid_invalid_format`
**Path:** `/api/users/by-uuid/not-a-valid-uuid`

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/validation/invalid-format",
  "title": "Invalid Format",
  "status": 422,
  "detail": "Parameter 'uuid' has invalid format. Expected UUID format",
  "instance": "/api/users/by-uuid/not-a-valid-uuid",
  "extensions": [
    {
      "parameter": "uuid",
      "expected_type": "uuid",
      "received_value": "not-a-valid-uuid",
      "location": "path"
    }
  ]
}
```

#### 4. Missing Required Parameter (Status: 422)

**Test:** `error_missing_required_query_param`
**Path:** `/api/users/search`

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/validation/missing-required-parameter",
  "title": "Missing Required Parameter",
  "status": 422,
  "detail": "Required parameter 'email' is missing",
  "instance": "/api/users/search",
  "extensions": [
    {
      "parameter": "email",
      "location": "query"
    }
  ]
}
```

#### 5. Endpoint Not Found (Status: 404)

**Test:** `error_404_unknown_endpoint`
**Path:** `/api/nonexistent`

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/endpoint-not-matched",
  "title": "Endpoint Not Found",
  "status": 404,
  "detail": "No endpoint matches the requested path",
  "instance": "/api/nonexistent"
}
```

**Note:** Generic HTTP errors (404, 405) may not include extensions if there's no specific parameter involved.

#### 6. Method Not Allowed (Status: 405)

**Test:** `error_405_wrong_method`
**Path:** `POST /api/users/1` (GET-only endpoint)

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/method-not-allowed",
  "title": "Method Not Allowed",
  "status": 405,
  "detail": "Method POST is not allowed for this endpoint",
  "instance": "/api/users/1"
}
```

**Note:** Generic HTTP errors (404, 405) may not include extensions if there's no specific parameter involved.

---

## 4. Test Data Reference

### Bootstrap SQL Schema

The test database is created from `test/test-data/bootstrap.sql` and contains:

**Tables:**
- `users` - User accounts (id, email, name, slug, uuid, active, rating, score, birth_date)
- `projects` - Projects (id, name, status, priority, owner_id, created_at)
- `tasks` - Tasks (id, title, details, status, priority, project_id, assignee_id, due_date, estimate, created_at)
- `measurements` - Sensor measurements (id, sensor_id, value, unit, precision_val, recorded_at)
- `logs` - Log entries (id, level, message, file_path, created_at)

**Sample Users:**
- Alice Carter (id=1, email=alice@example.com, active=1, score=85, rating=4.5)
- Bob Davis (id=2, email=bob@example.com, active=1, score=72, rating=4.8)
- Carol Evans (id=3, email=carol@example.com, active=0, score=50, rating=3.2)

**Sample Projects:**
- Website Redesign (id=1, status=active, priority=1, owner_id=1)
- Mobile App MVP (id=2, status=active, priority=2, owner_id=1)
- Data Migration (id=3, status=draft, priority=3, owner_id=2)

**Sample Tasks:**
- Design homepage (project_id=1, status=todo, priority=1)
- Implement responsive layout (project_id=1, status=todo, priority=2)
- API integration (project_id=2, status=doing, priority=1)

### Test Server Configuration

All tests use a root config structure that wraps individual endpoint definitions:

```json
{
  "$schema": "https://schemas.xmlui.org/v1/test-server/root-schema.json",
  "version": 1,
  "server": {
    "version": 1,
    "host": "127.0.0.1",
    "port": 8080,
    "api": {
      "version": 2,
      "name": "Test XMLUI Local Server API",
      "base_path": "/api",
      "webroot": ".",
      "endpoints": [
        // Individual endpoint config goes here
      ]
    }
  },
  "database": {
    "version": 1,
    "type": "sqlite3",
    "filepath": "test.db"
  }
}
```

---

## 5. Implementation Action Plan

### ✅ Phase 1: Foundation (COMPLETED)
- [x] Create test/TEST_PLAN.md
- [x] Create test/api_test_helpers.go (400+ lines)
- [x] Delete xmluisvr/apiresp/http_error_makers.go
- [x] Rename constraint `not-empty` → `notempty`
- [x] Fix regex constraint parser (use `strings.LastIndex()`)

### 🔄 Phase 2: Documentation Consolidation (IN PROGRESS)
- [ ] Create consolidated TEST_SPECIFICATION.md (this file)
- [ ] Verify all test scenarios documented
- [ ] Update ADR-011 with current status
- [ ] Archive old documentation files

### 📦 Phase 3: Extract Working Tests
**Goal:** Separate passing tests from failing/untested ones

- [ ] Create `test/api_datatypes_test.go`
  - Extract tests 02, 03, 04, 06, 07, 08, 14
  - Keep existing `shouldContain` validation
  - Verify all tests pass independently

- [ ] Create `test/api_constraints_test.go`
  - Extract tests 09, 10, 11
  - Document that validation is not enforced yet
  - Tests pass with valid inputs only

### 🔥 Phase 4: Fix Critical Blockers
**Priority:** CRITICAL (blocks most tests)

#### 4.1 Implement Parameter Type Validation
- [ ] Int type validator (reject "abc", "12.34")
- [ ] UUID format validator (standard 8-4-4-4-12 format)
- [ ] Slug format validator (lowercase, hyphens only)
- [ ] Date format validator (YYYY-MM-DD)
- [ ] Alphanumeric validator (no special chars)
- [ ] Boolean converter ("true"/"false" → 1/0)
- [ ] Return 422 with RFC 9457 on validation failure

#### 4.2 Implement Constraint Validation
- [ ] Range constraints (int, real, decimal)
- [ ] Length constraints (string length)
- [ ] Enum constraints (value in list)
- [ ] Regex constraints (pattern matching)
- [ ] NotEmpty constraint (non-empty string)
- [ ] Format constraints (date formats, UUID versions)
- [ ] Return 422 with ConstraintViolationErrorType

#### 4.3 Fix Parser Crashes
- [ ] Fix enum constraint parsing (handle commas)
- [ ] Fix regex constraint parsing (escape special chars)
- [ ] Implement UUID format constraint (`format[v4]`)

#### 4.4 Update Tests with RFC 9457 Validation
- [ ] Replace `shouldContain` with `expectedRFC9457` for all error tests
- [ ] Verify status codes are 422 (not 400) for validation errors
- [ ] Test all RFC 9457 fields exhaustively

### 📋 Phase 5: Implement Query Parameters
**Priority:** HIGH (blocks 6 tests)

- [ ] Implement query string parser
- [ ] Extract query parameters from URL
- [ ] Support optional parameters with defaults (`?{limit?10:int}`)
- [ ] Handle combined path + query parameters
- [ ] Create `test/api_query_params_test.go`
  - Extract tests 20, 21, 22
  - Verify query parameter extraction works
  - Test optional defaults

### 🚀 Phase 6: Advanced Features
**Priority:** MEDIUM (untested features)

- [ ] Create `test/api_advanced_test.go`
  - Test 23: Multi-segment parameters (`{file_path*}`)
  - Test 24: POST with JSON body parameters
  - Test 25: PUT with nested JSON body parameters

- [ ] Create `test/api_responses_test.go`
  - Test 27: Verify cardinality many returns array
  - Test 28: Verify row_type string returns plain text

- [ ] Create `test/api_errors_test.go`
  - Test 32: Verify 404 for unknown endpoints
  - Test 33: Verify 405 for wrong HTTP method
  - Update tests 29-31 with proper RFC 9457 validation

- [ ] Test 34: Complex scenario (all features combined)

### 🧹 Phase 7: Final Cleanup
**Priority:** LOW (quality improvements)

- [ ] Remove all `shouldContain` → use `expectedRFC9457` exclusively
- [ ] Delete `test/api_integration_test.go.old`
- [ ] Delete archived documentation (TEST_PLAN.md, UNIMPLEMENTED.md, RFC9457_TEST_REFERENCE.md)
- [ ] Update CLAUDE.md with new test structure
- [ ] Verify `make test` succeeds with all 34 scenarios passing

---

## 6. Success Criteria

**Project Complete When:**
- ✅ All 34 test scenarios passing
- ✅ All tests use `expectedRFC9457` validation (no `shouldContain`)
- ✅ Status codes correct (422 for validation, not 400)
- ✅ All RFC 9457 responses complete and accurate
- ✅ Critical blockers resolved (type validation, constraints, query params)
- ✅ Tests organized by category in separate files
- ✅ Documentation consolidated into TEST_SPECIFICATION.md
- ✅ Old test files deleted
- ✅ `make test` succeeds

---

## 7. Related Documents

- **ADR-007:** APIParamsMap Marshaling ([adrs/adr-007-api_params_map-unmarshalling.md](../adrs/adr-007-api_params_map-unmarshalling.md))
- **ADR-008:** SQL Placeholder Syntax ([adrs/adr-008-sql-parameters.md](../adrs/adr-008-sql-parameters.md))
- **ADR-009:** URL Arrays & Rows Syntax ([adrs/adr-009-multi-value-pathvars-syntax.md](../adrs/adr-009-multi-value-pathvars-syntax.md))
- **ADR-010:** Error Handling with RFC 9457 ([adrs/adr-010-error-handling-rfc9457.md](../adrs/adr-010-error-handling-rfc9457.md))
- **ADR-011:** Test Organization and RFC 9457 Implementation ([adrs/adr-011-test-organization-and-rfc9457-implementation.md](../adrs/adr-011-test-organization-and-rfc9457-implementation.md))
- **PathVars README:** ([xmluisvr/pathvars/README.md](../xmluisvr/pathvars/README.md))
- **PathVars ADR:** ([xmluisvr/pathvars/ADR_PATHVARS.md](../xmluisvr/pathvars/ADR_PATHVARS.md))

---

**Last Updated:** 2025-10-11
**Status:** Phase 2 (Documentation Consolidation) - IN PROGRESS