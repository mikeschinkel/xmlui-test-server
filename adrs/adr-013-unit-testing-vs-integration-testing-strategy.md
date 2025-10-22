# ADR-013: Unit Testing vs Integration Testing Strategy

|Label| Status                                                                                                              |
|--|---------------------------------------------------------------------------------------------------------------------|
|**Status:** | Pending                                                                                                                 |
|**Date:** | 2025-10-14                                                                                                          |
|**Related:** | [ADR-011: Test Organization and RFC 9457 Implementation](./adr-011-test-organization-and-rfc9457-implementation.md) |
|**Context:** | Clear separation needed between unit and integration tests across all packages                                      |

---

## Context and Problem Statement

The xmlui-test-server project has comprehensive integration tests in `./test/` that validate the full HTTP stack, and growing unit tests in various packages (`xmluisvr/pathvars/`, `xmluisvr/apiresp/`, `xmluisvr/cfgldr/`, etc.) that test package internals. However, we lacked clear, project-wide guidelines about:

1. **What belongs in unit tests vs integration tests?**
2. **When tests appear to test "the same thing", how do they differ?**
3. **How do we avoid duplication while maintaining comprehensive coverage?**
4. **What should each test layer verify?**
5. **Where should tests be located?**

### The Duplication Problem

Integration tests in `./test/` validate complete system behavior through HTTP requests and responses. Unit tests in package directories test internal logic, data structures, and error handling. Without clear boundaries, we risk:

- **Redundant testing** - Testing the same behavior at both layers
- **Missing coverage** - Gaps because each layer assumes the other tests something
- **Confusion** - Unclear which layer should test what
- **Slow feedback** - Over-reliance on integration tests slows development

**Central Question:** When the same scenario appears in both unit and integration tests, what distinguishes them?

---

## Decision Drivers

1. **Test execution speed** - Unit tests must be fast (~10 seconds or less per package)
2. **Test maintainability** - Clear purpose for each test prevents confusion
3. **Separation of concerns** - Each layer should test its own responsibilities
4. **Comprehensive coverage** - Both layers needed, but testing different aspects
5. **Developer productivity** - Fast feedback loop for package-level changes
6. **Scalability** - Strategy must work across all packages as project grows

---

## Considered Options

### Option 1: Only Integration Tests (Rejected)

**Approach:** Remove all unit tests, rely solely on comprehensive integration tests through HTTP.

**Pros:**
- ✅ No duplication of test cases
- ✅ Tests "real" behavior exactly as users experience it
- ✅ Simpler - one test strategy

**Cons:**
- ❌ Slow (requires full server startup, HTTP stack, database)
- ❌ Cannot test internal package structures
- ❌ Poor feedback loop during development
- ❌ Hard to test edge cases (requires complex HTTP scenarios)
- ❌ Cannot unit test internal logic and error composition

**Verdict:** Rejected - too slow and doesn't test package internals

---

### Option 2: Only Unit Tests (Rejected)

**Approach:** Test everything at package level, mock HTTP layer.

**Pros:**
- ✅ Fast execution
- ✅ Tests internal structures directly
- ✅ Easy to test edge cases

**Cons:**
- ❌ Misses integration issues (routing, HTTP status codes, JSON serialization)
- ❌ Doesn't validate API response formats
- ❌ Doesn't test real request/response cycle
- ❌ Requires extensive mocking of HTTP infrastructure

**Verdict:** Rejected - misses critical integration issues

---

### Option 3: Clear Separation with Defined Verification Boundaries (Chosen)

**Approach:** Both test layers exist, but each verifies different aspects of the system.

**Core Principle:**
> **Test cases CAN be shared or duplicated, but what you verify should differ based on the test layer.**

**Pros:**
- ✅ Fast unit tests for rapid development feedback
- ✅ Comprehensive integration tests for end-to-end validation
- ✅ Each layer has clear, distinct purpose
- ✅ Can test both internal logic AND external behavior
- ✅ Better failure diagnosis (know which layer failed)

**Cons:**
- ⚠️ More test code (same scenarios at different layers)
- ⚠️ Requires discipline to maintain separation
- ⚠️ Needs clear documentation and guidelines

**Verdict:** Chosen - balances speed, coverage, and maintainability

---

## Decision Outcome

### Core Principle

> **Test cases CAN be shared or duplicated, but what you verify should differ based on the test layer.**

This means:
- **Same scenario, different verifications** - Both layers may test "invalid input X", but unit tests verify internal error structures while integration tests verify HTTP responses
- **No redundancy in verification** - Each layer verifies aspects the other layer cannot or should not test
- **Complementary coverage** - Both layers together provide complete confidence

---

## Unit Tests (Package-Level)

### Definition

**Unit tests** test package internals directly without external dependencies.

### Location

```
xmluisvr/<package>/
  package.go
  package_test.go      ← Unit tests here (same directory as code)
```

**Examples:**
- `xmluisvr/pathvars/parsed_template_test.go`
- `xmluisvr/apiresp/response_payload_test.go`
- `xmluisvr/cfgldr/config_loader_test.go`
- `xmluisvr/common/rfc_9457_response_test.go`

### Characteristics

- **Direct function calls** - Call package functions/methods directly (no HTTP server)
- **Fast execution** - Entire package suite should run in ~10 seconds or less
- **Focused scope** - Individual functions, types, methods, internal logic
- **No external dependencies** - No database, no HTTP server, no file I/O (use mocks/stubs)
- **Can use `*http.Request` as input** - It's just a data structure; use `httptest.NewRequest()` to create test requests

### What Unit Tests SHOULD Verify

**Internal Data Structures:**
- ✅ Struct field population and initialization
- ✅ Type conversions and data transformations
- ✅ Internal state management

**Business Logic:**
- ✅ Validation rules and constraints
- ✅ Error composition and message generation
- ✅ Calculation and computation logic
- ✅ Conditional branching and decision trees

**Error Handling:**
- ✅ Error wrapping with `errors.Join()` and `errors.Wrap()`
- ✅ Error chain verification with `errors.Is()` and `errors.As()`
- ✅ Custom error types and fields

**Edge Cases:**
- ✅ Nil pointers, empty strings, zero values
- ✅ Boundary conditions (min/max values)
- ✅ Malformed input handling
- ✅ Unexpected data types

**Package Contracts:**
- ✅ Public API behavior and return values
- ✅ Interface implementations
- ✅ Type safety and compile-time guarantees

### What Unit Tests MUST NOT Verify

- ❌ HTTP status codes **in HTTP responses** (integration concern)
- ❌ HTTP response headers and body (integration concern)
- ❌ JSON/XML serialization to HTTP responses (integration concern)
- ❌ Database queries or connections (integration concern)
- ❌ File I/O operations (integration concern)
- ❌ Network communication (integration concern)
- ❌ Full request/response cycles over HTTP
- ❌ Server routing and middleware execution
- ❌ Cross-package integration behavior

**Important Distinction:**
- ✅ **CAN** use `*http.Request` as a data structure (via `httptest.NewRequest()`)
- ✅ **CAN** test functions that accept `*http.Request` as input
- ❌ **MUST NOT** start HTTP servers (`httptest.NewServer()`, `http.ListenAndServe()`)
- ❌ **MUST NOT** make network HTTP calls (`http.Get()`, `http.Client.Do()`)

### Example Pattern (Generic)

```go
// Package: xmluisvr/validator/
// File: validator_test.go

func TestValidator_InvalidInput(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        wantErr     error
        wantErrMsg  string
    }{
        {
            name:       "empty_input",
            input:      "",
            wantErr:    ErrEmptyInput,
            wantErrMsg: "input cannot be empty",
        },
        {
            name:       "invalid_format",
            input:      "invalid",
            wantErr:    ErrInvalidFormat,
            wantErrMsg: "expected format: ...",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validator := NewValidator()

            // Test internal package logic directly
            err := validator.Validate(tt.input)

            // Verify internal error structure
            require.Error(t, err)
            assert.True(t, errors.Is(err, tt.wantErr))
            assert.Contains(t, err.Error(), tt.wantErrMsg)
        })
    }
}
```

---

## Integration Tests (System-Level)

### Definition

**Integration tests** test the complete system through the HTTP layer, including all components working together.

### Location

```
test/
  api_datatypes_test.go       ← Integration tests for data type validation
  api_constraints_test.go     ← Integration tests for constraints
  api_parameters_test.go      ← Integration tests for parameter handling
  api_helpers_test.go         ← Shared test infrastructure
```

### Characteristics

- **Full HTTP stack** - Real HTTP requests and responses
- **Complete system** - Includes routing, parameter extraction, database, JSON serialization
- **Slower execution** - Requires server startup (acceptable for comprehensive coverage)
- **End-to-end behavior** - Validates the entire request/response cycle
- **External dependencies** - Database, configuration files, network stack

### What Integration Tests SHOULD Verify

**HTTP Layer:**
- ✅ HTTP status codes (200, 400, 404, 422, 500, etc.)
- ✅ HTTP headers (Content-Type, CORS, etc.)
- ✅ Request routing correctness

**Response Format:**
- ✅ JSON/XML structure and schema
- ✅ Response payload completeness
- ✅ Error response format (RFC 9457, custom formats)

**End-to-End Behavior:**
- ✅ Request parsing and validation
- ✅ Database query execution and results
- ✅ Authentication and authorization flows
- ✅ Real-world API usage scenarios

**Cross-Package Integration:**
- ✅ How packages work together
- ✅ Data flow through system layers
- ✅ Configuration loading and application

**System Contracts:**
- ✅ API contract compliance
- ✅ External interface correctness
- ✅ Backward compatibility

### What Integration Tests MUST NOT Verify

- ❌ Internal package structures (e.g., error structs)
- ❌ Package implementation details
- ❌ Error wrapping internals
- ❌ How errors are composed internally
- ❌ Private functions or unexported types

### Example Pattern (Generic)

```go
// Package: test
// File: api_validation_test.go

func TestAPI_InputValidation(t *testing.T) {
    server := setupTestServer(t)
    defer server.Cleanup()

    tests := []struct {
        name           string
        method         string
        path           string
        body           string
        expectedStatus int
        expectedJSON   map[string]any
    }{
        {
            name:           "invalid_parameter",
            method:         "GET",
            path:           "/api/resource/invalid",
            expectedStatus: 422,
            expectedJSON: map[string]any{
                "status": 422,
                "type":   "validation_error",
                "detail": "Invalid parameter format",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Make real HTTP request
            resp, body := makeRequest(server.BaseURL, tt.method, tt.path, tt.body)

            // Verify HTTP response
            assert.Equal(t, tt.expectedStatus, resp.StatusCode)

            // Verify JSON structure
            var result map[string]any
            json.Unmarshal(body, &result)
            assert.Equal(t, tt.expectedJSON["status"], result["status"])
            assert.Equal(t, tt.expectedJSON["type"], result["type"])
            assert.Contains(t, result["detail"], "Invalid parameter")
        })
    }
}
```

---

## Guidelines for Developers

### When Writing Unit Tests:

**DO:**
- ✅ Test package functions/methods directly
- ✅ Verify internal data structures and fields
- ✅ Test business logic and validation rules
- ✅ Test edge cases and error conditions
- ✅ Verify error wrapping with `errors.Is()` and `errors.As()`
- ✅ Keep tests fast (no HTTP, no database, minimal I/O)
- ✅ Focus on package-level correctness
- ✅ Test one package at a time in isolation

**DON'T:**
- ❌ Test HTTP responses or status codes **over the network**
- ❌ Test JSON/XML serialization to HTTP responses
- ❌ Test database queries
- ❌ Start HTTP servers (`httptest.NewServer()`)
- ❌ Make network HTTP requests (`http.Get()`, `http.Post()`)
- ❌ Test cross-package integration
- ❌ Test implementation details of other packages

**Important:** Using `*http.Request` as a function parameter is allowed—it's just a struct. Use `httptest.NewRequest()` to create test requests.

### When Writing Integration Tests:

**DO:**
- ✅ Test through HTTP layer (real requests/responses)
- ✅ Verify HTTP status codes and headers
- ✅ Verify response format (JSON, XML, etc.)
- ✅ Test end-to-end behavior
- ✅ Test real-world API scenarios
- ✅ Test cross-package interactions
- ✅ Include database and external dependencies

**DON'T:**
- ❌ Inspect internal package structures
- ❌ Test package implementation details
- ❌ Test error wrapping details
- ❌ Call package functions directly (use HTTP)
- ❌ Test without full system stack
- ❌ Mock HTTP layer (use real server)

---

## Test Naming and Organization

### Unit Test Files

**Location:** Same directory as source code
**Naming:** `<source_file>_test.go`

```
xmluisvr/validator/
  validator.go           ← Source code
  validator_test.go      ← Unit tests
  helper.go              ← Helper functions
  helper_test.go         ← Helper unit tests
```

### Integration Test Files

**Location:** `./test/` directory
**Naming:** `<feature_area>_test.go`

```
test/
  api_datatypes_test.go      ← Data type validation tests
  api_constraints_test.go    ← Constraint validation tests
  api_authentication_test.go ← Auth flow tests
  api_helpers_test.go        ← Shared test infrastructure
```

---

## Sharing Test Case Data (Optional)

While not required, test case data CAN be shared between layers if beneficial:

```go
// In xmluisvr/validator/test_cases.go (optional)
package validator

type ValidationTestCase struct {
    Name          string
    Input         string
    ExpectedError bool
    ErrorType     error
    ErrorMessage  string
}

var SharedValidationCases = []ValidationTestCase{
    {
        Name:          "empty_input",
        Input:         "",
        ExpectedError: true,
        ErrorType:     ErrEmptyInput,
        ErrorMessage:  "input cannot be empty",
    },
    // ... more cases
}
```

**Benefits:**
- Single source of truth for test scenarios
- Ensures unit and integration tests cover same cases
- Easier to add new test scenarios

**Considerations:**
- Not required - duplication is acceptable
- Integration tests may need additional fields (HTTP status, response body)
- Unit tests may test more edge cases not relevant to integration
- Each package decides if sharing is beneficial

---

## Positive Consequences

1. **Faster Development Feedback**
   - Unit tests run in ~10 seconds or less per package
   - Catch errors immediately during package development
   - No need to start server for every change

2. **Better Failure Diagnosis**
   - Unit test failure → package logic problem
   - Integration test failure → integration/HTTP layer problem
   - Clear separation helps identify root cause quickly

3. **Comprehensive Coverage**
   - Unit tests validate internal logic and edge cases
   - Integration tests validate end-to-end behavior
   - Both layers together provide complete confidence

4. **Clear Test Purposes**
   - No confusion about "why does this test exist?"
   - Each test has single, well-defined responsibility
   - Easier for new developers to understand

5. **Maintainable Test Suite**
   - Changes to internal structures only affect unit tests
   - Changes to API responses only affect integration tests
   - Reduced coupling between test layers

6. **Scalability**
   - Strategy works for any package in the project
   - Clear guidelines for all developers
   - Consistent approach across codebase

---

## Negative Consequences

1. **More Test Code**
   - Same scenarios tested at multiple layers
   - More lines of test code to maintain
   - **Mitigation:** Use test case sharing pattern where appropriate

2. **Requires Discipline**
   - Developers must understand layer boundaries
   - Need to resist testing everything at both layers
   - **Mitigation:** Clear guidelines in this ADR, code review enforcement

3. **Learning Curve**
   - New developers must understand unit vs integration distinction
   - May initially write tests at wrong layer
   - **Mitigation:** Examples in package-specific ADRs, mentoring during code review

---

## Validation Criteria

### Unit Test Suite (Per Package):
- ✅ Runs in ~10 seconds or less
- ✅ No HTTP servers or network calls
- ✅ No database dependencies
- ✅ Tests internal structures only
- ✅ Focuses on edge cases and error handling
- ✅ Fast feedback loop
- ✅ Can use `*http.Request` as input data structure

### Integration Test Suite:
- ✅ Tests through HTTP layer
- ✅ Validates API response formats
- ✅ Tests end-to-end scenarios
- ✅ Covers all API endpoints
- ✅ Includes database and external dependencies

### Both:
- ✅ No test verifies both internal structures AND HTTP responses
- ✅ Each test has single, clear purpose
- ✅ Test names clearly indicate what is being tested
- ✅ Clear separation maintained in code reviews

---

## Package-Specific Testing Strategies

This ADR provides the general strategy. Package-specific testing strategies are documented in separate ADRs:

- **[ADR-014: Testing Strategy for pathvars Package](./adr-014-testing-strategy-for-pathvars-package.md)** - How to test parameter validation and error handling
- **Future ADRs** - Additional packages will get specific testing strategies as needed:
  - `xmluisvr/apiresp/` - Response payload generation
  - `xmluisvr/cfgldr/` - Configuration loading and validation
  - `xmluisvr/common/` - Shared utilities and types

---

## Related Decisions

- **[ADR-011: Test Organization and RFC 9457 Implementation](./adr-011-test-organization-and-rfc9457-implementation.md)** - Overall test organization strategy
- **[ADR-010: Error Handling with RFC 9457](./adr-010-error-handling-rfc9457.md)** - RFC 9457 error response format
- **[ADR-014: Testing Strategy for pathvars Package](./adr-014-testing-strategy-for-pathvars-package.md)** - Package-specific testing strategy

---

## Implementation Plan

1. ✅ Document general strategy (this ADR)
2. ⏳ Create package-specific testing strategies (ADR-014, etc.)
3. ⏳ Review existing tests for compliance
4. ⏳ Refactor tests that violate layer boundaries
5. ⏳ Add missing unit tests for packages
6. ⏳ Update CLAUDE.md with testing guidelines
7. ⏳ Enforce in code reviews

---

**Last Updated:** 2025-10-14
**Status:** Pending (awaiting review and approval)
**Next Review:** After package-specific testing strategies are created
