
# ADR-014: Testing Strategy for pathvars Package

|Label| Status                                                                                                       |
|--|--------------------------------------------------------------------------------------------------------------|
|**Status:** | Pending                                                                                                          |
|**Date:** | 2025-10-14                                                                                                   |
|**Related:** | [ADR-013: Unit Testing vs Integration Testing Strategy](./adr-013-unit-testing-vs-integration-testing-strategy.md) |
|**Context:** | Specific testing strategy for xmluisvr/pathvars package                                                     |

---

## Context and Problem Statement

The `xmluisvr/pathvars` package is responsible for:
1. Parsing URL path templates (e.g., `/users/{id:int}/posts/{slug:string}`)
2. Matching incoming HTTP requests against templates
3. Extracting and validating path and query parameters
4. Generating detailed `ParameterValidationError` structures
5. Providing error information for RFC 9457 responses

**The testing challenge:** This package sits at the boundary between domain logic (parameter validation) and HTTP concerns (RFC 9457 responses). Without clear guidelines, we risk:

- **Duplication** - Testing RFC 9457 format in both unit and integration tests
- **Missing coverage** - Not testing internal error structures adequately
- **Confusion** - Unclear whether to test `ParameterValidationError` or RFC 9457 `Response`

**Specific example:** `TestTemplate_RFC9457_DetailMessage` currently tests only ONE template with minimal verification. Meanwhile, `./test/api_datatypes_test.go` comprehensively tests all data types through HTTP. What should the unit test actually test?

---

## Decision Drivers

1. **Adherence to ADR-013** - Follow general unit vs integration testing strategy
2. **Internal structure verification** - Need to test `ParameterValidationError` fields
3. **Error composition testing** - Need to test message generation functions
4. **Fast feedback** - Unit tests must run quickly during pathvars development
5. **Comprehensive coverage** - Cover all data types, constraints, and error scenarios
6. **No RFC 9457 duplication** - Integration tests already verify RFC 9457 responses

---

## Decision Outcome

### Core Principle (from ADR-013)

> **Test cases CAN be shared or duplicated, but what you verify should differ based on the test layer.**

### Application to pathvars Package

**Unit tests** (`xmluisvr/pathvars/*_test.go`):
- ✅ Test `ParameterValidationError` struct field population
- ✅ Test error composition functions (`createTypeValidationError`, `createConstraintViolationError`)
- ✅ Test error wrapping with `errors.Is()` and `errors.As()`
- ✅ Test `FaultSource` field (domain-level error classification)
- ✅ Can use `*http.Request` as input (via `httptest.NewRequest()`)
- ❌ Do NOT test RFC 9457 `Response` format (that's `apiresp` package's job)
- ❌ Do NOT test HTTP status codes in responses

**Integration tests** (`./test/api_datatypes_test.go`, etc.):
- ✅ Test RFC 9457 `Response` structure through HTTP
- ✅ Test HTTP status codes in HTTP responses
- ❌ Do NOT inspect `ParameterValidationError` internals
- ❌ Do NOT inspect `FaultSource` (that's internal to pathvars)

---

## Unit Testing Strategy

### What to Test

#### 1. `ParameterValidationError` Structure

**File:** `xmluisvr/pathvars/parsed_template_test.go`

**Test:** Comprehensive table-driven test covering all data types and error scenarios.

**Verify:**
```go
var pve *pathvars.ParameterValidationError
require.True(t, errors.As(err, &pve))

// Verify struct fields
assert.Equal(t, expectedParameter, pve.Parameter)
assert.Equal(t, expectedType, pve.ExpectedType)
assert.Equal(t, expectedValue, pve.ReceivedValue)
assert.Equal(t, expectedLocation, pve.Location)
assert.Contains(t, pve.Detail, expectedDetailFragment)
assert.Contains(t, pve.Suggestion, expectedSuggestionFragment)

// Verify internal fields (not exposed via RFC 9457)
assert.Equal(t, pathvars.ClientFault, pve.FaultSource)  // Internal field (not HTTP status!)
assert.Equal(t, template, pve.EndpointTemplate)
```

**Test Cases (Examples):**
- All data types: integer, string, UUID, slug, boolean, date, real, decimal, alphanumeric, identifier, email
- Path parameters vs query parameters
- Required vs optional parameters
- Single constraint failures
- Multiple constraint failures
- Type validation failures
- Constraint validation failures (when type is valid)

#### 2. Error Composition Functions

**File:** `xmluisvr/pathvars/error_composition_test.go` (new file)

**Functions to test:**
- `createTypeValidationError(validationArgs)` - Tests message composition for type errors
- `createConstraintViolationError(Constraint, validationArgs)` - Tests message composition for constraint errors

**Verify:**
```go
func TestCreateTypeValidationError(t *testing.T) {
    // Test that createTypeValidationError properly composes:
    // - Detail message format
    // - Suggestion message format
    // - Struct field population
    // - Error wrapping with ErrInvalidParameter
}

func TestCreateConstraintViolationError(t *testing.T) {
    // Test that createConstraintViolationError properly composes:
    // - Detail message from constraint.DetailMessage()
    // - Suggestion message from constraint.SuggestionMessage()
    // - Fallback to default messages
    // - Struct field population
}
```

#### 3. Error Wrapping and Unwrapping

**File:** `xmluisvr/pathvars/error_wrapping_test.go` (new file)

**Verify:**
```go
func TestErrorWrapping(t *testing.T) {
    tmpl, _ := pathvars.ParseTemplate("/api/users/{id:integer}")
    _, _, err := tmpl.Match("/api/users/abc", "")

    // Verify error chain contains expected sentinel errors
    assert.True(t, errors.Is(err, pathvars.ErrInvalidParameter))
    assert.True(t, errors.Is(err, pathvars.ErrInvalidParameterValue))

    // Verify ParameterValidationError can be extracted
    var pve *pathvars.ParameterValidationError
    assert.True(t, errors.As(err, &pve))

    // Verify error chain structure
    assert.NotNil(t, pve.Err)  // Wrapped error
}
```

#### 4. Constraint-Specific Unit Tests

**Files:** `xmluisvr/pathvars/constraint_*_test.go` (existing)

**Already covered:**
- `constraint_range_test.go` - Range constraint parsing and validation
- `constraint_uuid_format_test.go` - UUID format constraint
- `constraints_test.go` - General constraint testing

**Continue testing:**
- Constraint parsing logic
- Constraint validation logic
- Constraint error messages
- Constraint type detection

#### 5. Template Parsing and Matching

**Files:** `xmluisvr/pathvars/template_parsing_test.go`, `comprehensive_test.go` (existing)

**Continue testing:**
- Template parsing correctness
- Parameter extraction from paths
- Query parameter extraction
- Multi-segment parameters
- Optional parameters with defaults

---

## Integration Testing Strategy

### What to Test

Integration tests in `./test/` should focus on RFC 9457 responses through HTTP.

#### Test Files

**`./test/api_datatypes_test.go`** - Data type validation through HTTP
- ✅ Verify HTTP status 422 for invalid types
- ✅ Verify RFC 9457 `Response.Type` = `InvalidURLParameterErrorType`
- ✅ Verify RFC 9457 `Response.Title` = "Invalid URL Parameter"
- ✅ Verify RFC 9457 `Response.Detail` contains expected message
- ✅ Verify RFC 9457 `Response.Instance` = request path
- ✅ Verify RFC 9457 `Extensions` fields (Parameter, ExpectedType, ReceivedValue, Location, Suggestion)
- ❌ Do NOT inspect `ParameterValidationError` internals

**`./test/api_constraints_test.go`** - Constraint validation through HTTP
- ✅ Verify HTTP status 422 for constraint violations
- ✅ Verify RFC 9457 `Response.Type` = `ConstraintViolationErrorType`
- ✅ Verify RFC 9457 `Response.Detail` contains constraint info
- ✅ Verify RFC 9457 `Extensions` includes constraint details
- ❌ Do NOT test constraint parsing logic (that's unit test concern)

---

## Specific Test Examples

### Example 1: Unit Test for ParameterValidationError

```go
// File: xmluisvr/pathvars/parsed_template_test.go
// Purpose: Test internal ParameterValidationError structure

func TestParameterValidationError_AllDataTypes(t *testing.T) {
    tests := []struct {
        name          string
        template      string
        path          string
        query         string
        wantParameter string
        wantType      string
        wantValue     string
        wantLocation  string
        detailContains    string
        suggestionContains string
    }{
        {
            name:          "integer_type_invalid_string",
            template:      "/api/users/{id:integer}",
            path:          "/api/users/abc",
            wantParameter: "id",
            wantType:      "integer",
            wantValue:     "abc",
            wantLocation:  "path",
            detailContains:    "expected an integer",
            suggestionContains: "Use an integer",
        },
        {
            name:          "uuid_type_invalid_format",
            template:      "/api/items/{uuid:uuid}",
            path:          "/api/items/not-a-uuid",
            wantParameter: "uuid",
            wantType:      "uuid",
            wantValue:     "not-a-uuid",
            wantLocation:  "path",
            detailContains:    "expected a uuid",
            suggestionContains: "Use a uuid",
        },
        {
            name:          "query_param_invalid_boolean",
            template:      "/api/settings",
            path:          "/api/settings",
            query:         "enabled=maybe",
            wantParameter: "enabled",
            wantType:      "boolean",
            wantValue:     "maybe",
            wantLocation:  "query",
            detailContains:    "expected a boolean",
            suggestionContains: "Use a boolean",
        },
        // ... more test cases for all data types
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Parse template
            tmpl, err := pathvars.ParseTemplate(tt.template)
            require.NoError(t, err)

            // Attempt match (should fail validation)
            _, _, err = tmpl.Match(tt.path, tt.query)
            require.Error(t, err)

            // Extract ParameterValidationError
            var pve *pathvars.ParameterValidationError
            require.True(t, errors.As(err, &pve),
                "Expected ParameterValidationError in error chain")

            // Verify struct fields
            assert.Equal(t, tt.wantParameter, pve.Parameter)
            assert.Equal(t, tt.wantType, pve.ExpectedType)
            assert.Equal(t, tt.wantValue, pve.ReceivedValue)
            assert.Equal(t, tt.wantLocation, pve.Location)

            // Verify message composition
            assert.Contains(t, pve.Detail, tt.detailContains)
            assert.Contains(t, pve.Suggestion, tt.suggestionContains)
            assert.Contains(t, pve.Suggestion, "for example:")

            // Verify internal fields (domain concepts, NOT HTTP status codes!)
            assert.Equal(t, pathvars.ClientFault, pve.FaultSource)
            assert.Equal(t, tt.template, pve.EndpointTemplate)

            // Verify error wrapping
            assert.True(t, errors.Is(err, pathvars.ErrInvalidParameter))
            assert.True(t, errors.Is(err, pathvars.ErrInvalidParameterValue))
        })
    }
}
```

### Example 2: Integration Test for RFC 9457 Response

```go
// File: ./test/api_datatypes_test.go
// Purpose: Test RFC 9457 response through HTTP

func TestAPIDataTypes(t *testing.T) {
    server := setupComprehensiveTestServer(t)
    defer server.Cleanup()

    tests := []testRequest{
        {
            name:           "01_basic_int_parameter_invalid_string",
            method:         "GET",
            path:           "/api/users/abc",
            expectedStatus: 422,
            expectedRFC9457: &rfc9457.Response{
                Type:     rfc9457.InvalidURLParameterErrorType,
                Title:    "Invalid URL Parameter",
                Status:   422,
                Detail:   "Parameter 'id' expected an integer type but got 'abc'",
                Instance: "/api/users/abc",
                Extensions: []rfc9457.Extension{
                    apiresp.RFC9457Extension{
                        Parameter:     "id",
                        ExpectedType:  "integer",
                        ReceivedValue: "abc",
                        Location:      apiresp.PathLocation,
                        Suggestion:    "Use an integer for 'id' like 123, for example: /api/users/123",
                    },
                },
            },
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runTestRequest(t, server.BaseURL, tt)
        })
    }
}
```

**Notice:** The integration test verifies RFC 9457 response format, while the unit test verifies `ParameterValidationError` structure. Same scenario, different verification layers.

---

## Refactoring Plan

### Current State

**`TestTemplate_RFC9457_DetailMessage`** (lines 204-236 of `parsed_template_test.go`):
- ❌ Tests only ONE template (`/api/users/{id:integer}`)
- ❌ Minimal verification (checks if Detail contains certain strings)
- ❌ Not table-driven
- ❌ Doesn't cover all data types or scenarios

### Refactoring Steps

1. **Rename and expand:**
   - Rename: `TestTemplate_RFC9457_DetailMessage` → `TestParameterValidationError_AllDataTypes`
   - Convert to comprehensive table-driven test
   - Add all data types and scenarios

2. **Add new test files:**
   - `error_composition_test.go` - Test error message generation functions
   - `error_wrapping_test.go` - Test error chain structure

3. **Update existing tests:**
   - Ensure no existing pathvars unit tests verify RFC 9457 `Response` format
   - Focus all unit tests on `ParameterValidationError` and internal logic

4. **Verify integration tests:**
   - Confirm `./test/api_datatypes_test.go` does NOT inspect `ParameterValidationError`
   - Confirm integration tests focus only on RFC 9457 HTTP responses

---

## Test Coverage Goals

### Unit Tests

**Data Type Validation:**
- ✅ All data types: integer, string, UUID, slug, boolean, date, real, decimal, alphanumeric, identifier, email
- ✅ Valid values for each type
- ✅ Invalid values for each type
- ✅ Edge cases (empty, nil, boundary values)

**Constraint Validation:**
- ✅ Range constraints (int, real, decimal, date)
- ✅ Length constraints (string)
- ✅ Enum constraints
- ✅ Regex constraints
- ✅ NotEmpty constraints
- ✅ Format constraints (date, UUID)
- ✅ Multiple constraints on same parameter

**Parameter Locations:**
- ✅ Path parameters
- ✅ Query parameters
- ✅ Optional vs required parameters
- ✅ Parameters with default values

**Error Composition:**
- ✅ Type validation error messages
- ✅ Constraint violation error messages
- ✅ Detail message format
- ✅ Suggestion message format
- ✅ Error wrapping structure

### Integration Tests

**HTTP Responses:**
- ✅ All data types return 422 for invalid values
- ✅ RFC 9457 response structure complete
- ✅ RFC 9457 extensions populated correctly
- ✅ Suggestion messages helpful and accurate

**End-to-End:**
- ✅ Request routing works correctly
- ✅ Parameter extraction from path
- ✅ Parameter extraction from query string
- ✅ Database queries with extracted parameters

---

## Dependencies and Related Code

### Packages Involved

**`xmluisvr/pathvars/`** (this package):
- Creates `ParameterValidationError`
- Populates error struct fields
- Generates Detail and Suggestion messages
- Wraps errors with sentinel errors

**`xmluisvr/apiresp/`** (response package):
- Extracts `ParameterValidationError` using `errparsr`
- Maps `FaultSource` to HTTP status code
- Converts to RFC 9457 `Response`
- Serializes to JSON

**`./test/`** (integration tests):
- Makes HTTP requests
- Verifies RFC 9457 responses
- Does NOT inspect pathvars internals

### Separation of Concerns

```
pathvars package responsibilities:
├── Parse templates
├── Match requests
├── Validate parameters
├── Create ParameterValidationError ← Unit tests verify this
└── Populate error fields ← Unit tests verify this

apiresp package responsibilities:
├── Extract ParameterValidationError
├── Map FaultSource → HTTP status code ← Integration tests verify this
├── Convert to RFC 9457 Response ← Integration tests verify this
└── Serialize to JSON ← Integration tests verify this
```

---

## Validation Criteria

### Unit Tests (`xmluisvr/pathvars/*_test.go`):
- ✅ Cover all data types
- ✅ Cover all constraint types
- ✅ Test `ParameterValidationError` field population (including `FaultSource`)
- ✅ Test error message composition
- ✅ Test error wrapping
- ✅ Run in ~10 seconds or less
- ✅ Can use `*http.Request` as input data
- ✅ No HTTP servers or network calls
- ✅ No RFC 9457 `Response` verification

### Integration Tests (`./test/api_datatypes_test.go`, etc.):
- ✅ Cover all data types through HTTP
- ✅ Verify RFC 9457 response format
- ✅ Verify HTTP status codes
- ✅ Verify complete response structure
- ✅ No `ParameterValidationError` inspection

---

## Related Decisions

- **[ADR-013: Unit Testing vs Integration Testing Strategy](./adr-013-unit-testing-vs-integration-testing-strategy.md)** - General testing strategy
- **[ADR-011: Test Organization and RFC 9457 Implementation](./adr-011-test-organization-and-rfc9457-implementation.md)** - Test file organization
- **[ADR-010: Error Handling with RFC 9457](./adr-010-error-handling-rfc9457.md)** - RFC 9457 error format

---

## Implementation Plan

1. ✅ Document pathvars testing strategy (this ADR)
2. ⏳ Refactor `TestTemplate_RFC9457_DetailMessage` into comprehensive table-driven test
3. ⏳ Create `error_composition_test.go` to test message generation functions
4. ⏳ Create `error_wrapping_test.go` to test error chain structure
5. ⏳ Review all existing pathvars unit tests for compliance
6. ⏳ Verify integration tests do NOT inspect `ParameterValidationError`
7. ⏳ Add any missing unit test coverage
8. ⏳ Document examples in test files

---

## Lessons Learned

### Critical: errors.As() Requires Pointer Types

During integration testing implementation (2025-10-15), a critical bug was discovered that caused 13 tests to fail with incorrect 404 responses instead of 422 validation errors.

**Root Cause:**
```go
// ❌ WRONG - No pointer
var pve pathvars.ParameterValidationError
if errors.As(err, &pve) {
    // This never executes - errors.As() fails silently
}
```

**Fix:**
```go
// ✅ CORRECT - Pointer type
var pve *pathvars.ParameterValidationError
if errors.As(err, &pve) {
    // Now correctly extracts error
}
```

**Technical Explanation:**

The `errors.As()` function requires a pointer to populate the target variable:
```go
func As(err error, target any) bool
```

The `target` parameter must be a pointer to a type implementing `error`, or a pointer to a pointer to a concrete error type. When you pass `&pve` where `pve` is `pathvars.ParameterValidationError` (non-pointer), you're passing a pointer to a struct value. `errors.As()` cannot populate this correctly and returns `false`.

**Impact:**
- Without pointer type, `errors.As()` extraction fails silently
- Router continues searching for matching routes
- Eventually returns 404 "Endpoint Not Matched" instead of 422 "Invalid URL Parameter"
- Single-character bug (`*`) caused 13 test failures

**Testing Implication:**

This bug highlights why comprehensive integration tests are critical:
1. Unit tests of pathvars passed (error creation worked)
2. Unit tests with `errors.As()` would have caught this (if written correctly)
3. Integration tests revealed the bug through HTTP-level behavior

**Best Practice:**

Always use pointer types with `errors.As()`:
```go
// Extracting custom errors
var pve *pathvars.ParameterValidationError
var httpErr *httperr.HTTPError
var dbErr *database.QueryError

// All require pointer types
```

This lesson reinforces ADR-014's principle: **Integration tests catch bugs that unit tests miss**, especially architectural bugs in error handling pipelines.

---

**Last Updated:** 2025-10-15
**Status:** Accepted (amended with errors.As() lesson learned)
**Next Review:** After test refactoring is complete
