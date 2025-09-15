# ADR-006: JSON Versioning

**Status:** Accepted  
**Date:** 2025-09-15 
**Authors:** Claude, Mike Schinkel <mike@newclarity.net>

---

## 1. Context

XMLUI test server uses multiple JSON configuration types (API definitions, database configuration, server configuration) that require versioning for schema evolution. We need a consistent, universal property name for version identification that follows industry standards and supports controlled schema evolution across all configuration types.

---

## 2. Decision

Adopt `version` as the universal property name for all JSON configuration versioning with the following requirements:

- **Property name**: `version` (no prefix)
- **Data type**: Integer only (not string)
- **Scope**: Universal across all configuration schemas (API, database, server)
- **Validation**: Exact integer matching in JSON schemas

**Example**:
```json
{
  "$schema": "https://schemas.xmlui.org/v2/api-schema.json",
  "version": 2,
  "name": "Example API configuration",
  "base_path": "/api",
  "endpoints": [
     { "@note": "endpoints go here" }
  ]
}
```

---

## 3. Rationale

### 3.1 Industry Standard Alignment

The `version` property name follows established patterns across major platforms:
- **Kubernetes APIs**: Use `apiVersion` for API versioning
- **OpenAPI Specification**: Uses `version` in info objects
- **NPM packages**: Use `version` in package.json
- **Docker Compose**: Uses `version` at root level
- **GitHub Actions**: Use `version` in workflow definitions

### 3.2 Intentional Deployment Friction

Integer constraint creates deliberate barriers to schema evolution:
- **High friction**: Forces careful consideration before version changes
- **Complexity prevention**: Avoids overwhelming compatibility matrix management
- **Thoughtful evolution**: Encourages consolidating changes into meaningful releases
- **Simple comparison**: Integer comparison is deterministic and unambiguous

### 3.3 Universal Application

Single property name enables:
- **Consistent tooling**: Same validation and migration logic across all config types
- **Developer experience**: No confusion about versioning property names
- **Schema consistency**: Uniform approach across API, database, and server configurations

### 3.4 Backward Compatible Changes

Certain changes are considered backward compatible and do NOT require a version bump:

**Allowed without version change:**
- **Adding optional properties**: New fields that have default values or are not required
- **Loosening validation**: Making constraints less restrictive (e.g., increasing string length limits)
- **Adding new enum values**: Extending choice lists (with proper default handling)

**Requires version bump (breaking changes):**
- **Removing properties**: Deleting existing fields
- **Renaming properties**: Changing field names  
- **Type changes**: Modifying data types of existing fields
- **Making fields required**: Converting optional to required properties
- **Semantic changes**: Altering the meaning or behavior of existing properties
- **Tightening validation**: Making constraints more restrictive

**Example of safe addition:**

Version 2 - Original:
```json
{
  "version": 2,
  "name": "API Config",
  "endpoints": [
     { "@note": "endpoints go here" }
  ]
}
```

Version 2 - With added optional `description` property _(no version bump needed)_
```json
{
  "version": 2,
  "name": "API Config", 
  "description": "Optional description",  
   "endpoints": [
      { "@note": "endpoints go here" }
   ]
}
```

This approach follows JSON Schema evolution best practices where unknown properties are typically ignored by parsers, ensuring old code continues to function with newer configuration files.

---

## 4. Alternatives Considered

### 4.1 String Versioning (Semantic Versioning)
- **Rejected**: Low friction enables frequent breaking changes
- **Risk**: Leads to complex compatibility matrices (1.0.0, 1.1.0, 1.2.0, 2.0.0-beta, etc.)
- **Complexity**: String parsing and comparison logic vs simple integer comparison

### 4.2 Domain-Specific Names
- **`schemaVersion`**: Rejected as non-standard and verbose
- **`apiVersion`**: Rejected as domain-specific, not universal
- **`configVersion`**: Rejected as non-standard terminology

### 4.3 Prefixed Names
- **`$version`**: Rejected as non-standard (no JSON Schema precedent)
- **`_version`**: Rejected as implementation detail convention

---

## 5. Implementation

### 5.1 Schema Updates
- JSON schemas validate `version` as required integer with exact constants
- Struct tags updated from custom properties to `version`
- All v2+ configurations use standardized property

### 5.2 Backward Compatibility
- V1 configurations retain existing property names during transition
- Migration logic handles property name evolution between versions
- No breaking changes to existing v1 deployments

---

## 6. Consequences

### 6.1 Benefits
- **Standards compliance**: Follows widely accepted industry naming conventions
- **Controlled breaking changes**: High friction for breaking changes prevents schema sprawl and compatibility complexity
- **Safe evolution**: Allows backward compatible additions without version coordination overhead
- **Universal consistency**: Single property name across all configuration types
- **Developer clarity**: Clear guidelines on what requires version bumps vs safe additions
- **Simple tooling**: Integer comparison and validation logic

### 6.2 Trade-offs
- **Breaking change inflexibility**: Cannot use semantic versioning for gradual rollouts of breaking changes
- **Breaking change consolidation**: Must batch multiple breaking changes into single version bumps
- **Migration complexity**: Major version changes require comprehensive migration logic
- **Validation overhead**: Must carefully distinguish backward compatible vs breaking changes

---

## 7. Related ADRs

- **ADR-001**: Configuration Schema and Directory Structure (defines version detection)
- **ADR-002**: Load-Validate-Migrate Strategy (uses version for migration pipeline)
- **ADR-005**: API Configuration v2 (implements version property standardization)