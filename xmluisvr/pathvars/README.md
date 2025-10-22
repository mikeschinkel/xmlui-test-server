# pathvars

Path variable routing and parameter extraction for HTTP requests with typed parameters and validation constraints.

## Overview

The `pathvars` package provides path variable routing and parameter extraction functionality for HTTP requests. It supports typed path parameters with validation constraints, query parameter handling, and regex-based route matching.

Key features:
- Typed path parameters (string, int, uuid, date, etc.)
- Parameter validation constraints (range, length, regex, enum, etc.)
- Multi-segment parameters for capturing multiple path segments
- Optional parameters with default values
- Query parameter extraction and validation
- Efficient regex-based route matching

## Quick Start

```go
package main

import (
    "net/http"

    "xmlui-test-server/xmluisvr/pathvars"
)

func main() {
    router := pathvars.NewRouter()

    // Add a route with typed parameters
    params := []pathvars.Parameter{
        // Parameter definitions...
    }
    err := router.AddRoute("GET", "/users/{id:int}", params)
    if err != nil {
        // handle error
    }

    // Compile the router for efficient matching
    err = router.Compile()
    if err != nil {
        // handle error
    }

    // Match incoming requests
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        result, err := router.Match(r)
        if err == nil {
            userId, found := result.GetValue("id")
            // use extracted parameter
        }
    })
}
```

## Public API Reference

### Core Types

#### Router

The main routing engine that compiles and matches routes.

```go
type Router struct {
    // Contains private fields
}
```

**Functions:**
- `NewRouter() *Router` - Creates a new router instance
- `(r *Router) AddRoute(HTTPMethod, URLPath, pathvars.RouteArgs{...}) error` - Adds a route to the router
- `(r *Router) Compile() error` - Pre-compiles all routes for efficient matching
- `(r *Router) Match(*http.Request) (pathvars.MatchResult, error)` - Matches HTTP request against compiled routes

#### PathSpec, Method, Path

Type aliases for path specifications and components.

```go
type PathSpec string  // e.g., "GET /users/{id}" or "/users/{id}"
type Method string    // HTTP method like "GET", "POST"
type Path string      // URL path like "/users/{id}"
```

**Functions:**
- `ParsePathSpec(spec PathSpec) (method string, path string, err error)` - Splits path specification into method and path

#### Route

Represents a compiled HTTP endpoint with method, template, and routing index.

```go
type Route struct {
    Method   string    // HTTP method (empty = any method)
    Template *Template // Parsed path template
    Index    int       // Position in router's route list
}
```

#### Segment

Represents individual parts of a path template (literal strings or parameter placeholders).

```go
type Segment string
```

**Methods:**
- `(s Segment) IsLiteral() bool` - Returns true if segment is literal string (not parameter)
- `(s Segment) IsParameter() bool` - Returns true if segment is parameter placeholder

#### Parameter

Represents a path or query parameter with type and validation rules.

```go
type Parameter struct {
    // Contains private fields
}
```

**Creation:**
- `NewParameter(args ParameterArgs) Parameter` - Creates a new parameter instance
- `ParseParameter(spec string, position int) (Parameter, error)` - Parses parameter from specification string

**Methods:**
- `(p Parameter) DataType() PVDataType` - Returns parameter's data type
- `(p Parameter) Name() string` - Returns parameter name
- `(p Parameter) IsOptional() bool` - Returns true if parameter is optional
- `(p Parameter) IsMultiSegment() bool` - Returns true if parameter spans multiple path segments
- `(p Parameter) DefaultValue() *string` - Returns default value if any

**Configuration struct:**
```go
type ParameterArgs struct {
    Name         string
    UseType      ParamUseType
    DataType     PVDataType
    Constraints  []Constraint
    Position     int
    Original     string
    MultiSegment bool
    Optional     bool
    DefaultValue *string
}
```

#### ParamUseType

Indicates how a parameter is used.

```go
type ParamUseType int

const (
    UnspecifiedParameterType ParamUseType = iota
    PathParameter    // Extracted from URL path
    QueryParameter   // Extracted from query string
)
```

#### Template

Represents a parsed path template with parameters and compiled regex.

```go
type Template struct {
    // Contains private fields
}
```

**Creation:**
- `ParseTemplate(template string) (*Template, error)` - Parses template string into Template object

**Methods:**
- `(t *Template) Match(path, queryString string) (ValuesMap, bool)` - Matches path and query against template
- `(t *Template) Parameters() []Parameter` - Returns all parameters (TODO: implementation needed)
- `(t *Template) Validate(params map[string]string) error` - Validates parameter values (TODO: implementation needed)
- `(t *Template) Substitute(values map[string]string) (string, error)` - Builds path from values (TODO: implementation needed)

#### MatchResult

Contains the results of matching an HTTP request against routes.

```go
type MatchResult struct {
    Index int // Which route matched
    // Contains private fields
}

type ValuesMap map[string]string
```

**Creation:**
- `NewMatchResult(index int, valuesMap ValuesMap) MatchResult` - Creates new match result

**Methods:**
- `(m MatchResult) ParamsMap() ValuesMap` - Returns extracted parameter values
- `(m MatchResult) GetValue(name string) (value string, found bool)` - Gets specific parameter value
- `(m MatchResult) VarCount() int` - Returns number of extracted parameters
- `(m MatchResult) HasVars() bool` - Returns true if any parameters were extracted
- `(m MatchResult) ForEachVar(fn func(name, value string) bool)` - Iterates over parameters

### Data Types

#### PVDataType

Enumerated data types for parameter validation.

```go
type PVDataType int

const (
    UnspecifiedType PVDataType = iota
    StringType
    IntegerType
    RealType
    DecimalType
    IdentifierType
    DateType
    UUIDType
    AlphanumericType
    SlugType
    BooleanType
    EmailType
)
```

**Methods:**
- `(dt PVDataType) TypeName() PVDataTypeName` - Returns canonical string name

**Functions:**
- `ParsePVDataType(typeStr string) (PVDataType, error)` - Converts string to data type
- `InferDataTypeFromName(name string) (PVDataType, bool)` - Infers type from parameter name

#### PVDataTypeName

String representation of data types.

```go
type PVDataTypeName string

const (
    StringTypeName       PVDataTypeName = "string"
    IntegerTypeName      PVDataTypeName = "integer"
    IntTypeName          PVDataTypeName = "int"        // Alias for integer
    DecimalTypeName      PVDataTypeName = "decimal"
    RealTypeName         PVDataTypeName = "real"
    IdentifierTypeName   PVDataTypeName = "identifier"
    DateTypeName         PVDataTypeName = "date"
    UUIDTypeName         PVDataTypeName = "uuid"
    AlphanumericTypeName PVDataTypeName = "alphanumeric"
    AlphanumTypeName     PVDataTypeName = "alphanum"   // Alias for alphanumeric
    SlugTypeName         PVDataTypeName = "slug"
    BooleanTypeName      PVDataTypeName = "boolean"
    BoolTypeName         PVDataTypeName = "bool"       // Alias for boolean
    EmailTypeName        PVDataTypeName = "email"
)
```

### Constraints

#### Constraint Interface

Defines parameter validation constraints.

```go
type Constraint interface {
    Validate(value string) error
    String() string
    Type() ConstraintType
    Parse(value string, dataType PVDataType) (Constraint, error)
    ValidDateTypes() []PVDataType
    MapKey(dt PVDataTypeName) ConstraintMapKey
    EnsureBaseConstraint(Constraint)
}
```

#### ConstraintType

Types of validation constraints.

```go
type ConstraintType string

const (
    FormatConstraintType    ConstraintType = "format"
    EnumConstraintType      ConstraintType = "enum"
    LengthConstraintType    ConstraintType = "length"
    NotEmptyConstraintType  ConstraintType = "notempty"
    RangeConstraintType     ConstraintType = "range"
    RegexConstraintType     ConstraintType = "regex"
)
```

**Functions:**
- `ParseConstraints(spec string, dataType PVDataType) ([]Constraint, error)` - Parses constraint specifications

#### Constraint Registry

Functions for managing constraint types and data type aliases.

```go
type ConstraintMapKey string
type ConstraintsMap map[ConstraintMapKey]Constraint
type DataTypeAliasMap = map[PVDataTypeName]PVDataTypeName
```

**Functions:**
- `RegisterDataTypeAlias(dataType PVDataType, alias PVDataTypeName)` - Registers type alias
- `RegisterConstraint(c Constraint)` - Registers a constraint implementation
- `GetConstraintsMap() ConstraintsMap` - Returns the global constraints map
- `GetConstraintMapKey(ct ConstraintType, dtn PVDataTypeName) ConstraintMapKey` - Generates constraint key
- `GetConstraint(ct ConstraintType, dt PVDataType) (Constraint, error)` - Retrieves constraint by type

#### Specific Constraint Types

The package provides several built-in constraint implementations:

**DateFormatConstraint:**
```go
type DateFormatConstraint struct { /* private fields */ }
```
- `NewDateFormatConstraint(format string, parser func(string) (time.Time, error)) *DateFormatConstraint`
- `ParseDateFormatConstraint(spec string) (*DateFormatConstraint, error)`

**DateRangeConstraint:**
```go
type DateRangeConstraint struct { /* private fields */ }
```
- `NewDateRangeConstraint(min time.Time, max time.Time) *DateRangeConstraint`
- `ParseDateRangeConstraint(rangeSpec string) (*DateRangeConstraint, error)`

**DecimalRangeConstraint:**
```go
type DecimalRangeConstraint struct { /* private fields */ }
```
- `NewDecimalRangeConstraint(min float64, max float64) *DecimalRangeConstraint`
- `ParseDecimalRangeConstraint(rangeSpec string) (*DecimalRangeConstraint, error)`

**EnumConstraint:**
```go
type EnumConstraint struct { /* private fields */ }
```
- `NewEnumConstraint(values map[string]bool, list []string) *EnumConstraint`
- `ParseEnumConstraint(enumSpec string) (*EnumConstraint, error)`

**IntegerRangeConstraint:**
```go
type IntegerRangeConstraint struct { /* private fields */ }
```
- `NewIntRangeConstraint(min int64, max int64) *IntegerRangeConstraint`
- `ParseIntRangeConstraint(rangeSpec string) (*IntegerRangeConstraint, error)`

**LengthConstraint:**
```go
type LengthConstraint struct { /* private fields */ }
```
- `NewLengthConstraint(min int, max int) *LengthConstraint`
- `ParseLengthConstraint(rangeSpec string) (*LengthConstraint, error)`

**NotEmptyConstraint:**
```go
type NotEmptyConstraint struct { /* private fields */ }
```
- `NewNotEmptyConstraint() *NotEmptyConstraint`
- `ParseNotEmptyConstraint(value string) (*NotEmptyConstraint, error)`

**RegexConstraint:**
```go
type RegexConstraint struct { /* private fields */ }
```
- `NewRegexConstraint(regex *regexp.Regexp, raw string) *RegexConstraint`
- `ParseRegexConstraint(pattern string) (*RegexConstraint, error)`

**UUIDFormatConstraint:**
```go
type UUIDFormatConstraint struct { /* private fields */ }
```
- `NewUUIDFormatConstraint(format string, validator func(string) error) *UUIDFormatConstraint`
- `ParseUUIDFormatConstraint(spec string) (*UUIDFormatConstraint, error)`

**Utility Functions:**
- `ParseRangeConstraint(rangeSpec string, dataType PVDataType) (Constraint, error)` - Generic range constraint parser

### Error Handling

The package defines several sentinel error values for different failure scenarios:

```go
var (
    ErrInvalidTemplate        = errors.New("invalid template syntax")
    ErrUnmatchedBrace         = errors.New("unmatched brace in template")
    ErrInvalidParameter       = errors.New("invalid parameter")
    ErrInvalidType            = errors.New("unknown parameter type")
    ErrInvalidConstraint      = errors.New("invalid constraint syntax")
    ErrNoMatch                = errors.New("no matching route")
    ErrAPIRouterNotCompiled   = errors.New("API router not compiled; must be compiled before calling Match()")
    ErrValidationFailed       = errors.New("parameter validation failed")
    ErrUnknownConstraintType  = errors.New("unknown constraint type")
    ErrInvalidSyntax          = errors.New("invalid syntax")
    ErrParseFailed            = errors.New("parse failed")
)
```

All errors provide detailed context including the failing value, expected format, and error location through error wrapping.

## Parameter Syntax

Parameters use a flexible syntax in path templates:

### Basic Syntax
- `{name}` - String parameter, type inferred from name if possible
- `{name:type}` - Explicit data type
- `{name:type:constraints}` - Type with validation constraints
- `{name::constraints}` - Inferred type with constraints (double colon)

### Optional Parameters
- `{name?}` - Optional parameter, no default
- `{name?default}` - Optional parameter with default value

### Multi-segment Parameters
- `{name*}` - Captures multiple path segments
- `{name*?}` - Optional multi-segment parameter

### Constraint Examples
- `{id:int:range[1..1000]}` - Integer between 1 and 1000
- `{email:string:regex[.+@.+]}` - String matching email pattern (auto-anchored for full match)
- `{status:string:enum[active,inactive]}` - String from allowed values
- `{name:string:length[3..50]}` - String with length constraints
- `{slug:string:notempty}` - Non-empty string
- `{date:date:format[yyyy-mm-dd]}` - Date with specific format

### Multiple Constraints
- `{id:string:regex[[0-9]+],length[3..10]}` - Multiple constraints separated by commas

**Note on Regex Constraints:** Regex patterns automatically match the complete parameter value (full string matching). Do not include `^` (start) or `$` (end) anchors in your patterns - they are added automatically to ensure security and prevent partial matches. For example, `regex[.+@.+]` internally becomes `^.+@.+$` before compilation.

## Usage Examples

### Simple Route
```go
router.AddRoute("GET", "/users/{id:int}", []Parameter{})
```

### Route with Constraints
```go
router.AddRoute("GET", "/users/{id:int:range[1..1000]}", []Parameter{})
```

### Route with Query Parameters
```go
params := []Parameter{
    NewParameter(ParameterArgs{
        Name:     "limit",
        UseType:  QueryParameter,
        DataType: IntegerType,
        Optional: true,
        DefaultValue: stringPtr("10"),
    }),
}
router.AddRoute("GET", "/users", params)
```

### Optional Parameters with Defaults
```go
router.AddRoute("GET", "/posts/{category?general:string}", []Parameter{})
```

### Multi-segment Parameters
```go
router.AddRoute("GET", "/files/{path*:string}", []Parameter{})
```

This README provides comprehensive documentation of all public APIs in the pathvars package, including types, functions, methods, constants, and usage examples.