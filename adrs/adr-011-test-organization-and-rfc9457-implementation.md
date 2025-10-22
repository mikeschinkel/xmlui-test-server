What# ADR-011: Test Organization and RFC 9457 Implementation

|Label|Status|
|--|--|
|**Status:** | In Progress|
|**Date:** | 2025-10-05|
|**Related:** | [ADR-010: Error Handling with RFC 9457](./adr-010-error-handling-rfc9457.md)|
|**Context:** | Test reorganization plan in [test/TEST_SPECIFICATION.md](../test/TEST_SPECIFICATION.md)|

---

## Context

The xmlui-test-server project has grown to 1,495 lines of integration tests in a single file (`test/api_integration_test.go`) with 84 test scenarios. Additionally, we are implementing RFC 9457 error responses throughout the system, which requires architectural decisions about error creation patterns, test organization, and code reusability.

**Problems to solve:**
1. **Large monolithic test file** - Hard to navigate, slow to run, difficult to maintain
2. **Missing unit tests** - No unit tests for RFC 9457 generation logic in domain packages
3. **Code duplication** - Repeated validation logic across tests
4. **Legacy test patterns** - Using fragile `shouldContain` string matching instead of structured RFC 9457 validation
5. **Shared type problem** - RFC9457Response needed across multiple packages (see `/Users/mikeschinkel/Documents/InfoWorld/shared-types-in-golang.md`)

---

## Decision 1: RFC9457Response Location and Instantiation Pattern

### Decision
**Keep RFC9457Response in `xmluisvr/common/` package and allow domain packages to instantiate directly.**

### Rationale
We considered two options:

**Option A (Chosen):** Instantiate RFC9457Response in domain packages
- ✅ Type-safe, context-aware error messages
- ✅ No information loss via error wrapping
- ✅ Works well with ParsedError extraction pattern
- ❌ Creates dependency on common package

**Option B (Rejected):** Instantiate only in apiresp
- ✅ Centralized error handling
- ❌ Requires duplicating context info via `fmt.Errorf()` metadata
- ❌ Less type-safe, more fragile parsing
- ❌ Loss of compile-time type safety

### Implementation Pattern

Domain packages create RFC9457Response and wrap in errors.Join():

```go
// In pathvars/template.go validateParameter()
err = doterr.NewErr(
    ErrInvalidParameter,
    ErrInvalidParameterValue,
    "template", t.raw,
    common.NewRFC9457Response(common.RFC9457ResponseArgs{
        Type:          common.InvalidParameterErrorType,
        Title:         "Invalid Parameter Type",
        Status:        http.StatusUnprocessableEntity,
        Detail:        detail,
        Instance:      source,
Extensions: []rfc9457.Extension{
apiresp.RFC9457Extension{

Parameter:     string(param.Name),
        ExpectedType:  string(param.dataType.Slug()),
        ReceivedValue: value,
        Location:      common.LocationType(location.Slug()),
    },
},
}	),
    err,
)
```

API handler extracts RFC9457Response using ParsedError:

```go
// In apipkg/api_handler.go
pe, _ := errparsr.ParseError(err)
err = pe.MaybeGetCustomError(common.RFC9457ResponseType)
if err != nil && errors.As(err, &rfc9457) {
    httpStatus = rfc9457.Status
    response.Send(apiresp.UnprocessableEntityPayload(r, rfc9457))
}
```

### Consequences
- ✅ Type-safe error creation at point of failure
- ✅ Context-rich error messages without string parsing
- ✅ Solves shared-types problem pragmatically (accept dependency on common)
- ⚠️ All packages using RFC9457Response depend on common
- ✅ ParsedError pattern (xmluisvr/errparsr) provides clean extraction

---

## Decision 2: Error Creation Pattern - Args Structs, Not Multi-Parameter Functions

### Decision
**Use Args struct pattern for all error creation. Reject multi-parameter functions (>3 parameters).**

### Rationale

**Anti-pattern (Rejected):**
```go
// ❌ BAD: 5 parameters is a code smell
func MakeInvalidParameterError(
    param, expectedType, receivedValue string,
    location LocationType,
    instance string,
) *RFC9457Response {
    return &RFC9457Response{
        Type:          InvalidParameterErrorType,
        Title:         "Invalid Parameter Type",
        Status:        422,
        Detail:        fmt.Sprintf("..."),
        Instance:      instance,
Extensions: []rfc9457.Extension{
apiresp.RFC9457Extension{

Parameter:     param,
        ExpectedType:  expectedType,
        ReceivedValue: receivedValue,
        Location:      location,
    },
},
}	
}
```

**Preferred pattern (Adopted):**
```go
// ✅ GOOD: Args struct with named fields
common.NewRFC9457Response(common.RFC9457ResponseArgs{
    Type:          common.InvalidParameterErrorType,
    Title:         "Invalid Parameter Type",
    Status:        422,
    Detail:        detail,
    Instance:      instance,
Extensions: []rfc9457.Extension{
apiresp.RFC9457Extension{

Parameter:     param,
    ExpectedType:  expectedType,
    ReceivedValue: receivedValue,
    Location:      location,
},
},
}	)
```

### Benefits
- ✅ Named fields - self-documenting
- ✅ Optional fields - can omit unused fields
- ✅ Type safety at compile time
- ✅ Easy to extend with new fields
- ✅ IDE autocomplete support
- ✅ Refactoring-friendly

### Consequences
- ✅ Deleted `xmluisvr/apiresp/http_error_makers.go` (used anti-pattern)
- ✅ All error creation follows consistent pattern
- ⚠️ Slightly more verbose at call site (acceptable tradeoff)

---

## Decision 3: Constraint Naming Convention

### Decision
**All constraint names must be lowercase alpha only. No hyphens, underscores, or special characters.**

### Specific Changes
- ✅ Renamed `not-empty` → `notempty`
- ✅ Future constraints follow pattern: `range`, `length`, `enum`, `regex`, `notempty`, `format`
- ✅ May add numeric characters later if needed (not currently required)

### Rationale
**Problem:** The constraint `not-empty` caused parser crashes due to hyphen character:
```
panic: invalid syntax, position=4, character=-, reason=invalid constraint type character
```

**Root cause:** Parser expected simple alpha identifiers, hyphen triggered edge case in parsing state machine.

**Fix:** Simplify naming convention to eliminate edge cases:
- Simpler parsing logic
- More consistent naming
- Eliminates entire class of parsing bugs

### Implementation
Updated in 7+ files:
- `xmluisvr/pathvars/constraints.go` - Constant definition
- `xmluisvr/pathvars/constraint_not_empty.go` - Comments
- `test/api_integration_test.go` - Test paths (4 occurrences)
- `test/test-data/api_comprehensive_test.json` - Test data
- `test/UNIMPLEMENTED.md` - Documentation
- `xmluisvr/pathvars/README.md` - Documentation
- `xmluisvr/pathvars/ADR_PATHVARS.md` - ADR update

### Consequences
- ✅ Fixes crash bug in tests 16, 17, 31, 34
- ✅ Simpler parser implementation
- ✅ More consistent naming across constraints
- ⚠️ Breaking change for existing configs using `not-empty` (acceptable - pre-1.0)

---

## Decision 4: Test Organization Strategy

### Decision
**Split tests into unit tests (side-by-side with code) and integration tests (in ./test), organized by feature category.**

### File Organization

```
xmluisvr/common/
  rfc_9457_response.go
  rfc_9457_response_test.go        ← Unit tests for RFC9457Response struct

xmluisvr/pathvars/
  template.go
  template_test.go                 ← Unit tests for validateParameter() RFC9457 generation
  validator.go
  validator_test.go                ← Unit tests for validation logic

xmluisvr/apiresp/
  response_payload.go
  response_payload_test.go         ← Unit tests for payload maker functions

test/
  api_test_helpers.go              ← Shared test infrastructure (NEW)

  # Integration tests (split by category)
  api_datatypes_test.go            ← Tests 01-08: int, string, uuid, etc.
  api_constraints_test.go          ← Tests 09-17: range, length, enum, etc.
  api_parameters_test.go           ← Tests 18-22: query params, implicit types
  api_advanced_test.go             ← Tests 23-25: multi-segment, JSON body
  api_responses_test.go            ← Tests 26-28: cardinality, row_type
  api_errors_test.go               ← Tests 29-33: 404, 405, 422, 500
  api_complex_test.go              ← Test 34: combined features

  api_integration_test.go          → DELETE after migration
```

### Rationale

**Problems with single 1,495-line file:**
- Hard to navigate (find specific test)
- Slow to run (all tests run together)
- Difficult to maintain (merge conflicts)
- No clear organization (all categories mixed)

**Benefits of split approach:**
- ✅ Each file independently runnable
- ✅ Organized by feature category
- ✅ Easier to navigate and maintain
- ✅ Faster targeted test runs
- ✅ Clear separation: unit vs integration
- ✅ Reduced merge conflicts

**Unit test placement rationale:**
- Tests for package-internal logic live with the code
- Makes package more self-contained
- Encourages testing during development
- Examples: RFC9457 generation, validation logic, payload makers

**Integration test placement rationale:**
- Tests that require full server stack stay in ./test
- Tests that cross package boundaries
- End-to-end API behavior tests

### Consequences
- ✅ Better organization and maintainability
- ✅ Faster targeted test execution
- ✅ Easier onboarding for new contributors
- ⚠️ More files to manage (acceptable tradeoff)
- ✅ Clear distinction between unit and integration tests

---

## Decision 5: Shared Test Infrastructure

### Decision
**Create `test/api_test_helpers.go` with reusable test infrastructure to eliminate duplication.**

### Implementation

Created 400+ line helper file with:

**RFC 9457 Assertion:**
```go
func assertRFC9457Equal(t *testing.T, got, want *common.RFC9457Response) {
    t.Helper()
    // Field-by-field comparison with clear error messages
    // Compares all 11 main fields + ValidationErrors slice
}
```

**Server Lifecycle Management:**
```go
type testServer struct {
    BaseURL      string
    Port         int
    ctx          context.Context
    cancel       context.CancelFunc
    wg           *sync.WaitGroup
    serverError  error
    env          *testEnvironment
    t            *testing.T
}

func setupTestServer(t *testing.T, testName, configContent string) *testServer
func (s *testServer) Cleanup()
func (s *testServer) runTestRequest(req testRequest)
```

**Test Environment Setup:**
```go
type testEnvironment struct {
    rootFixture        *fsfix.RootFixture
    dbPath             string
    bootstrapFile      *fsfix.FileFixture
    configFile         *fsfix.FileFixture
    configStoreMap     cfgstore.ConfigStoresMap
    bufferedLogHandler *testutil.BufferedLogHandler
    bufferedWriter     *testutil.BufferedWriter
    logger             *slog.Logger
}

func setupTestEnvironment(t *testing.T, testName, configContent string) *testEnvironment
func (env *testEnvironment) cleanup(t *testing.T)
```

**Utility Functions:**
```go
func makeHTTPRequest(baseURL, method, path, body string) (*http.Response, []byte, error)
func findAvailablePort() (int, error)
func getRootConfig(endpointJSON string) string
func closeOrError(t *testing.T, closer io.Closer)
```

### Rationale

**Before:** Each test file duplicated:
- Server setup/teardown (50+ lines)
- HTTP request execution (30+ lines)
- RFC 9457 validation (70+ lines)
- Environment configuration (40+ lines)

**After:** Each test file can:
```go
server := setupTestServer(t, "test_name", configJSON)
defer server.Cleanup()

server.runTestRequest(testRequest{
    name:            "test_case",
    method:          "GET",
    path:            "/api/users/abc",
    expectedStatus:  422,
    expectedRFC9457: &common.RFC9457Response{...},
})
```

### Benefits
- ✅ ~200 lines of duplicated code eliminated per test file
- ✅ Consistent test execution patterns
- ✅ Easier to maintain (fix bug in one place)
- ✅ Better test isolation (each test gets clean server)
- ✅ Improved readability (tests focus on what, not how)

### Consequences
- ✅ All integration tests can use shared infrastructure
- ✅ Reduced maintenance burden
- ✅ Consistent error reporting across all tests
- ⚠️ Learning curve for understanding helper abstractions (minor)

---

## Decision 6: Migration from `shouldContain` to `expectedRFC9457`

### Decision
**Replace all legacy `shouldContain` string matching with structured `expectedRFC9457` validation.**

### Current State

**Legacy pattern (being phased out):**
```go
testRequest{
    name:           "get_user_by_int_id_invalid_string",
    method:         "GET",
    path:           "/api/users/abc",
    expectedStatus: 422,
    shouldContain:  []string{"invalid", "parameter", "type"}, // ❌ Fragile
}
```

**New pattern (preferred):**
```go
testRequest{
    name:           "get_user_by_int_id_invalid_string",
    method:         "GET",
    path:           "/api/users/abc",
    expectedStatus: 422,
    expectedRFC9457: &common.RFC9457Response{              // ✅ Structured
        Type:          common.InvalidParameterErrorType,
        Title:         "Invalid Parameter Type",
        Status:        422,
        Detail:        "Parameter 'id' expected an integer type but got 'abc'",
        Instance:      "/api/users/abc",
Extensions: []rfc9457.Extension{
apiresp.RFC9457Extension{

Parameter:     "id",
        ExpectedType:  "integer",
        ReceivedValue: "abc",
        Location:      common.PathLocation,
    },
},
}	,
}
```

### Rationale

**Problems with `shouldContain`:**
- ❌ Fragile - breaks with message wording changes
- ❌ Incomplete - doesn't validate structure
- ❌ Ambiguous - substring matches can be misleading
- ❌ No type safety - strings can have typos

**Benefits of `expectedRFC9457`:**
- ✅ Validates complete RFC 9457 structure
- ✅ Type-safe - compiler catches errors
- ✅ Self-documenting - shows expected response shape
- ✅ Precise - validates all fields exhaustively
- ✅ Refactoring-friendly - compiler helps with changes

### Migration Strategy
1. Add `expectedRFC9457` field to test cases
2. Validate using `assertRFC9457Equal()` helper
3. Remove `shouldContain` after validation passes
4. Delete legacy validation code once all tests migrated

### Status
- ✅ Validation infrastructure in place (`assertRFC9457Equal`)
- ⏳ 3 of 84 test scenarios migrated to `expectedRFC9457`
- ⏳ 81 scenarios still using legacy `shouldContain`
- 📋 Tracked in test/TEST_SPECIFICATION.md Phase 4

### Consequences
- ✅ More robust test suite
- ✅ Better documentation of expected behavior
- ✅ Easier to maintain as RFC 9457 evolves
- ⚠️ More verbose test definitions (acceptable for clarity)

---

## Decisions Pending (To Be Updated)

### Areas Requiring Future Decisions

1. **Query Parameter Extraction**
   - How to extract query params from URL vs JSON body?
   - Where should extraction logic live?
   - Status: Not yet implemented (tests 20-22 failing)

2. **Constraint Validation Architecture**
   - Where should validation happen? (parser vs handler)
   - How to return multiple validation errors?
   - Status: Not yet implemented (tests 09-17 partial)

3. **Error Response Aggregation**
   - Should we collect all validation errors before responding?
   - Or fail-fast on first error?
   - Status: To be decided during implementation

4. **Test Data Management**
   - Should test data (SQL, configs) be inline or external files?
   - Current: Mix of inline and external
   - Status: To be standardized

*This section will be updated as implementation progresses and new decisions are made.*

---

## Implementation Status

### Phase 1: Foundation ✅ **COMPLETED**
- ✅ test/TEST_SPECIFICATION.md created (consolidated test plan)
- ✅ test/api_test_helpers.go created (400+ lines)
- ✅ xmluisvr/apiresp/http_error_makers.go deleted
- ✅ Constraint `not-empty` renamed to `notempty`

### Phase 2: Documentation Consolidation ✅ **COMPLETED (2025-10-11)**
- ✅ test/TEST_SPECIFICATION.md created (consolidated all test docs)
  - 34 test scenarios with complete status
  - RFC 9457 error response reference
  - Implementation action plan with priorities
  - Test data reference and bootstrap SQL documentation
- ✅ Updated ADR-011 with current status
- ⏳ Archival of old documentation files (pending)

### Phase 3: Extract Working Tests 📦 **PENDING**
- ⏳ test/api_datatypes_test.go (tests 02, 03, 04, 06, 07, 08, 14)
- ⏳ test/api_constraints_test.go (tests 09, 10, 11)
- ⏳ test/api_query_params_test.go (tests 20, 21, 22)
- ⏳ test/api_advanced_test.go (tests 23, 24, 25)
- ⏳ test/api_responses_test.go (tests 26, 27, 28)
- ⏳ test/api_errors_test.go (tests 29, 30, 31, 32, 33)
- ⏳ test/api_complex_test.go (test 34)

### Phase 4: Fix Critical Blockers 🔥 **PENDING**
- ⏳ Implement parameter type validation (int, uuid, slug, date, alphanumeric)
- ⏳ Implement constraint validation (range, length, enum, regex, notempty)
- ⏳ Fix parser crashes (enum commas, regex escaping, UUID format constraint)
- ⏳ Implement query parameter extraction
- ⏳ Implement boolean type converter
- ⏳ Update all tests with RFC 9457 validation

### Phase 5: Unit Tests (Package-level) 📦 **PENDING**
- ⏳ xmluisvr/pathvars/template_test.go
- ⏳ xmluisvr/apiresp/response_payload_test.go

### Phase 6: Final Cleanup 🧹 **PENDING**
- ⏳ Remove all `shouldContain` usage
- ⏳ Delete test/api_integration_test.go.old
- ⏳ Update CLAUDE.md with new test structure
- ⏳ Verify `make test` succeeds with all 34 scenarios passing

---

## Related Documents

- **[test/TEST_SPECIFICATION.md](../test/TEST_SPECIFICATION.md)** - ✅ **ACTIVE** - Consolidated test specification (all test scenarios, status, RFC 9457 reference, implementation plan)
- [ADR-010: Error Handling with RFC 9457](./adr-010-error-handling-rfc9457.md) - Original RFC 9457 decision
- [xmluisvr/pathvars/ADR_PATHVARS.md](../xmluisvr/pathvars/ADR_PATHVARS.md) - Path variables architecture

---

**Last Updated:** 2025-10-11
**Next Review:** After Phase 3 completion (test extraction)
