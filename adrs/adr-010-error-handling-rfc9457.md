# ADR-010: Error Handling with RFC 9457 Problem Details

* **Status:** Accepted
* **Date:** 2025-10-01
* **Authors:** Claude, Mike Schinkel <mike@newclarity.net>

---

## 1. Context

The XMLUI test server needs a comprehensive error handling strategy that:

1. **Validates API parameters** (types, constraints, formats) and returns clear errors
2. **Provides detailed error information** to help developers quickly diagnose and fix issues
3. **Follows HTTP standards** for status codes and error responses
4. **Maintains clean architectural separation** between reusable components and project-specific logic
5. **Offers excellent developer experience (DX)** through self-documenting, discoverable error responses

### 1.1 Problem Statement

Current error handling has several gaps:

- **No parameter validation**: Invalid types like `"abc"` for `{id:int}` are silently converted or cause unclear errors
- **Inconsistent error responses**: Error messages lack structure and vary in format
- **Poor DX**: Error responses don't provide enough context to fix issues
- **Missing validation layer**: Unclear where validation should occur (routing layer vs. application layer)
- **Non-resolvable error identifiers**: No standard way to identify and document error types

### 1.2 Requirements

1. Return **structured, machine-readable** error responses
2. Include **detailed context** (parameter name, expected type, received value, location)
3. Use **appropriate HTTP status codes** (not just 400/500 for everything)
4. Provide **resolvable URLs** that link to error documentation
5. Validate **in the correct architectural layer** (general framework vs. project-specific)
6. Support both **simple** (single parameter) and **complex** (multiple validation errors) scenarios

---

## 2. Decision

Adopt **RFC 9457 Problem Details for HTTP APIs** as the standard error response format, with custom extensions for parameter validation details.

### 2.1 HTTP Status Code Strategy

**Use 422 Unprocessable Entity** for all parameter validation errors:

| Status Code | Use Case | Example |
|-------------|----------|---------|
| **422** | Type validation failure | `"abc"` for `{id:int}` parameter |
| **422** | Constraint violation | `150` for `{score:int:range[0..100]}` |
| **422** | Format validation failure | `"not-a-uuid"` for `{uuid:uuid}` |
| **422** | Missing required parameter | Query parameter marked required but absent |
| 400 | Malformed request syntax | Invalid JSON body, malformed headers |
| 404 | Resource not found | Valid request but resource doesn't exist in database |
| 500 | Internal server error | Unexpected errors, database failures |

**Rationale for 422 over 400:**
- **400 Bad Request**: Syntax/parsing errors (malformed JSON, bad HTTP headers)
- **422 Unprocessable Entity**: Semantic validation errors (valid syntax, invalid values)
- **Industry standard**: GitHub, Stripe, Shopify, JSON:API all use 422 for validation
- **More precise**: Distinguishes between "can't parse your request" (400) vs. "understood your request but values are invalid" (422)

### 2.2 RFC 9457 Response Structure

**Base structure** (required fields):

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type",
  "title": "Invalid Parameter Type",
  "status": 422,
  "detail": "Parameter 'id' must be an integer, received '12.34'",
  "instance": "/api/users/12.34"
}
```

**With validation extensions** (custom fields):

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type",
  "title": "Invalid Parameter Type",
  "status": 422,
  "detail": "Parameter 'id' must be an integer, received '12.34'",
  "instance": "/api/users/12.34",

  "parameter": "id",
  "expected_type": "int",
  "received_value": "12.34",
  "location": "path",
  "constraint": null,
  "suggestion": "Use an integer like /api/users/12"
}
```

**For multiple validation errors**:

```json
{
  "type": "https://schema.xmlui.org/errors/test-server/api/validation/multiple-errors",
  "title": "Multiple Validation Errors",
  "status": 422,
  "detail": "Request contains 2 validation errors",
  "instance": "/api/users/search",

  "validation_errors": [
    {
      "parameter": "id",
      "location": "path",
      "expected": "int",
      "received": "abc",
      "message": "Must be an integer"
    },
    {
      "parameter": "email",
      "location": "query",
      "expected": "email format",
      "received": "notanemail",
      "message": "Must be valid email format"
    }
  ]
}
```

### 2.3 Error Type URLs

Use **`schema.xmlui.org`** domain for all error type URLs:

```
https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type
https://schema.xmlui.org/errors/test-server/api/validation/constraint-violation
https://schema.xmlui.org/errors/test-server/api/validation/missing-required-parameter
https://schema.xmlui.org/errors/test-server/api/validation/invalid-format
https://schema.xmlui.org/errors/test-server/api/database/query-failed
https://schema.xmlui.org/errors/test-server/api/database/no-results
```

**Rationale:**
- Namespace consistency with existing JSON schemas at `schema.xmlui.org`
- Semantic clarity: "schema" represents structured definitions (errors are structured)
- Single domain for all structured definitions (schemas + error types)
- Real-world precedent: `schema.org` uses same domain for multiple purposes

### 2.4 Architectural Decision: Validation Layer

**Parameter validation occurs in the `pathvars` package (routing layer), not `apipkg` (application layer).**

**Request flow:**

```
1. HTTP Request arrives
2. pathvars.Router.Match(r)           ← VALIDATION HAPPENS HERE
   - Match URL pattern
   - Extract parameter values
   - Validate types (int, uuid, slug, date, etc.)
   - Validate constraints (range, length, enum, regex)
   - Return: MatchResult OR httperr.HTTPError
3. apipkg handles result
   - If validation error: write RFC 9457 response
   - If success: extract query values and execute SQL
```

**Why validate in `pathvars`, not `apipkg`:**

| Aspect | Reasoning |
|--------|-----------|
| **Ownership** | `pathvars` defines the constraint DSL (`{id:int}`, `range[0..100]`), so it should enforce it |
| **Reusability** | Any project using `pathvars` gets validation for free |
| **Separation of concerns** | Routing layer: "Is this valid input?", Application layer: "What should I do with valid input?" |
| **Single source of truth** | Constraint semantics live where constraints are parsed |
| **Framework pattern** | Frameworks validate input before passing to application code |

**Code location:**

```go
// In api_handler.go, line 44:
result, err = api.Router.Match(r)

if errors.Is(err, pathvars.ErrNoMatch) {
    http.NotFound(w, r)
    goto end
}

// NEW: Handle validation errors from pathvars
var httpErr *httperr.HTTPError
if errors.As(err, &httpErr) {
    httpErr.Write(w)  // Send RFC 9457 error
    goto end
}

if err != nil {
    // Unexpected internal error
    api.internalServerError(result, w, r, err, "Unexpected error")
    goto end
}

// Continue with validated parameters...
```

### 2.5 Resolvable URI Philosophy

**All URIs MUST resolve to useful content for developers.**

This is a core DX principle: never use non-resolvable URIs like `urn:xmlui:error:validation:001`.

**Implementation:**

```
Directory structure at schema.xmlui.org:
└── errors/
    └── validation/
        ├── invalid-parameter-type/
        │   ├── index.html        ← Human documentation
        │   └── example.json      ← Machine-readable example
        ├── constraint-violation/
        │   ├── index.html
        │   └── example.json
        └── ...
```

**URLs:**

- `https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type` → HTML documentation (for humans)
- `https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type/example.json` → JSON example (for machines)

**Benefits:**
- Developers click error type URL → instant documentation
- Self-documenting errors reduce support burden
- Professional polish and attention to developer experience
- Discoverable: HTML page links to JSON example

---

## 3. Implementation

### 3.1 Go Struct Definitions

**File:** `xmluisvr/httperr/error.go`

```go
package httperr

// HTTPError represents an RFC 9457 Problem Details error response
type HTTPError struct {
    // RFC 9457 required fields
    Type     string `json:"type"`               // URI identifying error type
    Title    string `json:"title"`              // Short, human-readable summary
    Status   int    `json:"status"`             // HTTP status code
    Detail   string `json:"detail"`             // Human-readable explanation
    Instance string `json:"instance"`           // URI reference identifying occurrence

    // Custom extensions for parameter validation
    Parameter        string            `json:"parameter,omitempty"`         // Parameter name
    ExpectedType     string            `json:"expected_type,omitempty"`     // Expected data type
    ReceivedValue    string            `json:"received_value,omitempty"`    // Actual value received
    Location         string            `json:"location,omitempty"`          // "path", "query", or "body"
    Constraint       any               `json:"constraint,omitempty"`        // Constraint that failed
    Suggestion       string            `json:"suggestion,omitempty"`        // How to fix it
    ValidationErrors []ValidationError `json:"validation_errors,omitempty"` // Multiple errors
}

type ValidationError struct {
    Parameter string `json:"parameter"`
    Location  string `json:"location"`
    Expected  string `json:"expected"`
    Received  string `json:"received"`
    Message   string `json:"message"`
}
```

### 3.2 Helper Functions

**File:** `xmluisvr/httperr/helpers.go`

```go
package httperr

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// Error type URL constants
const (
    TypeInvalidParameterType   = "https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type"
    TypeConstraintViolation    = "https://schema.xmlui.org/errors/test-server/api/validation/constraint-violation"
    TypeMissingParameter       = "https://schema.xmlui.org/errors/test-server/api/validation/missing-required-parameter"
    TypeInvalidFormat          = "https://schema.xmlui.org/errors/test-server/api/validation/invalid-format"
    TypeQueryFailed            = "https://schema.xmlui.org/errors/test-server/api/database/query-failed"
    TypeNoResults              = "https://schema.xmlui.org/errors/test-server/api/database/no-results"
)

// Location constants
const (
    LocationPath  = "path"
    LocationQuery = "query"
    LocationBody  = "body"
)

// InvalidParameterType creates an error for type validation failures
func InvalidParameterType(param, expectedType, receivedValue, location, instance string) *HTTPError {
    return &HTTPError{
        Type:          TypeInvalidParameterType,
        Title:         "Invalid Parameter Type",
        Status:        http.StatusUnprocessableEntity, // 422
        Detail:        fmt.Sprintf("Parameter '%s' must be %s, received '%s'", param, expectedType, receivedValue),
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

// ConstraintViolation creates an error for constraint violations
func ConstraintViolation(param, constraint, receivedValue, location, instance string) *HTTPError {
    return &HTTPError{
        Type:          TypeConstraintViolation,
        Title:         "Constraint Violation",
        Status:        http.StatusUnprocessableEntity, // 422
        Detail:        fmt.Sprintf("Parameter '%s' violates constraint %s, received '%s'", param, constraint, receivedValue),
        Instance:      instance,
Extensions: []rfc9457.Extension{
apiresp.RFC9457Extension{

Parameter:     param,
        ReceivedValue: receivedValue,
        Location:      location,
        Constraint:    constraint,
    },
},
}	
}

// MissingParameter creates an error for missing required parameters
func MissingParameter(param, location, instance string) *HTTPError {
    return &HTTPError{
        Type:      TypeMissingParameter,
        Title:     "Missing Required Parameter",
        Status:    http.StatusUnprocessableEntity, // 422
        Detail:    fmt.Sprintf("Required parameter '%s' is missing from %s", param, location),
        Instance:  instance,
Extensions: []rfc9457.Extension{
apiresp.RFC9457Extension{

Parameter: param,
        Location:  location,
    },
},
}	
}

// Mutation methods (no return value for clearer semantics and easier debugging)
func (e *HTTPError) SetSuggestion(suggestion string) {
    e.Suggestion = suggestion
}

func (e *HTTPError) SetConstraint(constraint string) {
    e.Constraint = constraint
}

func (e *HTTPError) AddValidationError(ve ValidationError) {
    if e.ValidationErrors == nil {
        e.ValidationErrors = []ValidationError{}
    }
    e.ValidationErrors = append(e.ValidationErrors, ve)
}

// Write sends the error as an RFC 9457 JSON response
func (e *HTTPError) Write(w http.ResponseWriter) error {
    w.Header().Set("Content-Type", "application/problem+json") // RFC 9457 media type
    w.WriteHeader(e.Status)
    return json.NewEncoder(w).Encode(e)
}

// Internal creates a generic internal server error (don't expose internals)
func Internal(instance string, err error) *HTTPError {
    return &HTTPError{
        Type:     "https://schema.xmlui.org/errors/test-server/api/internal-server-error",
        Title:    "Internal Server Error",
        Status:   http.StatusInternalServerError,
        Detail:   "An unexpected error occurred",
        Instance: instance,
    }
}
```

### 3.3 Usage Examples

**Example 1: Simple type validation error**

```go
// In pathvars.Router.Match()
if !isValidInt(value) {
    return MatchResult{}, httperr.InvalidParameterType(
        "id",           // parameter name
        "int",          // expected type
        value,          // received value (e.g., "12.34")
        httperr.LocationPath,
        r.URL.Path,     // instance
    )
}
```

**Example 2: Constraint violation with suggestion**

```go
err := httperr.ConstraintViolation(
    "score",
    "range[0..100]",
    "150",
    httperr.LocationPath,
    r.URL.Path,
)
err.SetSuggestion("Use a value between 0 and 100, like /api/users/by-score/85")
return MatchResult{}, err
```

**Example 3: Multiple validation errors**

```go
err := &httperr.HTTPError{
    Type:     httperr.TypeInvalidParameterType,
    Title:    "Multiple Validation Errors",
    Status:   422,
    Detail:   "Request contains multiple validation errors",
    Instance: r.URL.Path,
}
err.AddValidationError(ValidationError{
    Parameter: "id",
    Location:  "path",
    Expected:  "int",
    Received:  "abc",
    Message:   "Must be an integer",
})
err.AddValidationError(ValidationError{
    Parameter: "email",
    Location:  "query",
    Expected:  "email format",
    Received:  "notanemail",
    Message:   "Must be valid email",
})
return MatchResult{}, err
```

### 3.4 Test Expectations

**Updated test structure:**

```go
type testRequest struct {
    name           string
    method         string
    path           string
    body           string
    expectedStatus int    // 422 for validation errors
    expectedFields []string
    shouldContain  []string  // e.g., []string{"invalid", "int"}
    shouldNotHave  []string
}

// Example test case
{
    name:           "get_user_by_int_id_invalid_float",
    method:         "GET",
    path:           "/api/users/12.34",
    expectedStatus: 422,  // Changed from 400
    shouldContain:  []string{"invalid", "int"},
},
```

### 3.5 Error Documentation

**Documentation structure at `schema.xmlui.org`:**

Each error type has:
1. **`index.html`**: Human-readable documentation with:
   - Error description
   - Common causes
   - How to fix
   - Examples (valid/invalid)
   - Related errors
   - Link to `example.json`

2. **`example.json`**: Machine-readable example response

**HTML template structure:**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Invalid Parameter Type - XMLUI Error Reference</title>
</head>
<body>
    <h1>Invalid Parameter Type</h1>
    <p><span class="status-code">422 Unprocessable Entity</span></p>

    <a href="example.json" class="example-link">📄 View JSON Example</a>

    <h2>Description</h2>
    <p>This error occurs when a URL parameter value does not match its declared data type...</p>

    <h2>Example Response</h2>
    <pre><code>{ ... JSON example ... }</code></pre>

    <h2>Common Causes</h2>
    <ul>
        <li>Passing "abc" when an integer is expected</li>
        <li>Passing 12.34 when an integer is required</li>
    </ul>

    <h2>How to Fix</h2>
    <ol>
        <li>Check the API documentation for expected types</li>
        <li>Ensure parameter values match declared types</li>
    </ol>

    <h2>Related Errors</h2>
    <ul>
        <li><a href="../constraint-violation/">Constraint Violation</a></li>
    </ul>
</body>
</html>
```

---

## 4. Consequences

### 4.1 Benefits

✅ **Standardized error format**: Follows RFC 9457 industry standard
✅ **Better DX**: Detailed errors help developers fix issues quickly
✅ **Resolvable URIs**: Every error type URL links to documentation
✅ **Precise status codes**: 422 clearly indicates validation errors
✅ **Clean architecture**: Validation in correct layer (pathvars, not apipkg)
✅ **Reusable**: pathvars package provides validation for any project
✅ **Extensible**: Easy to add new error types and fields
✅ **Machine-readable**: Structured format enables automated error handling
✅ **Self-documenting**: Clicking error URLs provides instant help

### 4.2 Trade-offs

⚠️ **More verbose responses**: RFC 9457 errors are larger than simple strings
⚠️ **Documentation maintenance**: Need to keep error docs up-to-date
⚠️ **Implementation complexity**: More sophisticated than simple error strings
⚠️ **Breaking change**: Changes existing error response format

### 4.3 Mitigations

- **Response size**: Gzip compression reduces JSON overhead
- **Documentation**: Generate docs from code/templates to reduce manual maintenance
- **Implementation**: Helper functions make creation simple (one-liners)
- **Migration**: Phase in gradually, starting with new endpoints

### 4.4 Future Considerations

**Phase 1** (Current):
- Basic structure (type, title, status, detail, instance)
- Core validation fields (parameter, expected_type, received_value, location)
- Simple error constructor functions

**Phase 2** (Future):
- Multiple validation errors in single response
- More sophisticated constraint error messages
- Localized error messages (i18n)

**Phase 3** (Future):
- Error analytics and tracking
- Client SDKs that parse RFC 9457 errors
- Auto-generated API client code with type checking

---

## 5. Alternatives Considered

### 5.1 Alternative: Use 400 Instead of 422

**Rejected because:**
- Less precise: 400 is for malformed requests (syntax), not invalid values (semantics)
- Industry trend: Modern APIs use 422 for validation
- HTTP spec supports it: RFC 9110 defines 422 for semantic validation

### 5.2 Alternative: Simple Error Strings

```json
{ "error": "Invalid parameter 'id': expected int, got '12.34'" }
```

**Rejected because:**
- Not machine-readable: Hard to parse programmatically
- No standard structure: Every error format is different
- Missing context: Can't easily extract parameter name, expected type, etc.
- No discoverability: No way to link to documentation

### 5.3 Alternative: Custom Error Format

Create our own error structure instead of RFC 9457.

**Rejected because:**
- Reinventing the wheel: RFC 9457 already solves this
- No industry recognition: Developers familiar with RFC 9457
- Missing best practices: RFC incorporates years of experience
- Harder to integrate: Tools/libraries support RFC 9457

### 5.4 Alternative: Validate in `apipkg` Layer

Put validation in `GetParameterValues()` (line 60) instead of `Router.Match()`.

**Rejected because:**
- Wrong ownership: `apipkg` doesn't define constraint DSL
- Coupling: Makes `apipkg` coupled to validation logic
- Not reusable: Other projects using `pathvars` don't get validation
- Framework pattern: Routing layers should validate before passing to app

### 5.5 Alternative: Non-resolvable URIs

Use URNs like `urn:xmlui:error:validation:invalid-parameter-type`.

**Rejected because:**
- Poor DX: Developers can't click to get help
- Extra work: Must search docs separately
- No discoverability: No automatic way to find documentation
- Against philosophy: We want all URIs to be helpful resources

---

## 6. References

### 6.1 Standards

- **[RFC 9457](https://www.rfc-editor.org/rfc/rfc9457)**: Problem Details for HTTP APIs (July 2023)
  - Obsoletes RFC 7807 (March 2016)
  - Same core structure, refined after real-world usage

- **[RFC 9110](https://httpwg.org/specs/rfc9110.html)**: HTTP Semantics
  - Section 15.5.23: 422 Unprocessable Entity

- **[RFC 3986](https://www.rfc-editor.org/rfc/rfc3986)**: Uniform Resource Identifier (URI): Generic Syntax

### 6.2 Industry Examples

- **GitHub API**: Uses 422 for validation errors with structured responses
- **Stripe API**: Uses 422 with detailed error objects
- **Shopify API**: Uses 422 for validation failures
- **JSON:API Specification**: Mandates 422 for validation errors

### 6.3 Related ADRs

- **ADR-001**: Configuration Schema and Directory Structure
- **ADR-002**: Versioning Schemas
- **ADR-005**: API Config
- **ADR-009**: Multi-Value PathVars Syntax

### 6.4 External Resources

- [Problem Details for HTTP APIs (RFC 9457)](https://www.rfc-editor.org/rfc/rfc9457)
- [HTTP Status Code 422](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/422)
- [Schema.org](https://schema.org) - Example of resolvable URIs for structured data
- [JSON:API Error Handling](https://jsonapi.org/format/#errors)

---

## 7. Implementation Checklist

- [ ] Create `xmluisvr/httperr` package with error structs and helpers
- [ ] Define error type URL constants
- [ ] Implement validation in `pathvars.Router.Match()`
- [ ] Update `api_handler.go` to handle `httperr.HTTPError`
- [ ] Update test expectations to expect 422 status code
- [ ] Create error documentation site structure at `schema.xmlui.org/errors/test-server/api/`
- [ ] Write documentation for each error type (HTML + JSON examples)
- [ ] Add breadcrumb navigation to error docs
- [ ] Test error responses match RFC 9457 format
- [ ] Validate `Content-Type: application/problem+json` header is set

---

## 8. Notes

**Date:** 2025-10-01

**Key Insight:** Treating `pathvars` as a reusable framework component (rather than project-specific code) naturally leads to the correct architectural decision: validation belongs in the layer that defines and parses the constraint DSL.

**Developer Experience Philosophy:** All URIs should resolve to useful content. Non-resolvable URIs are inconsiderate to developers and create unnecessary friction. The small effort to provide documentation at error type URLs pays massive dividends in developer happiness and reduced support burden.

**Progressive Implementation:** Start with basic error structure and core validation types. Add more sophisticated features (multiple errors, suggestions, localization) as needs emerge from real usage.
