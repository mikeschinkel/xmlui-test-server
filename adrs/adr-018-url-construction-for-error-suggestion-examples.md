# ADR-018: URL Construction for Error Suggestion Examples

**Status**: Accepted
**Date**: 2025-10-20
**Context**: RFC 9457 error responses with suggestion URLs

---

## Context

When API parameter validation fails, we provide RFC 9457 error responses that include suggestion URLs showing users how to fix their request. The construction and format of these suggestion URLs significantly impacts usability - they must be clear, scannable, and guide attention to the problematic parameter.

### User Experience Goals

1. **Minimize cognitive load** - Users should understand the suggestion at a glance
2. **Guide attention** - The problematic parameter should be immediately obvious
3. **Avoid confusion** - Don't show values that might be mistaken for literal requirements
4. **Maintain scannability** - Verbose actual values can obscure the pattern
5. **Stay focused** - Only show parameters relevant to the user's actual request

### Initial Considerations

We evaluated several approaches for formatting parameter values:

1. **Use actual user values throughout** - `?email=invalid-email&min_score=50`
   - ❌ Verbose values make patterns harder to see
   - ❌ Creates cognitive overhead ("which value is wrong?")
   - ❌ May leak sensitive data in logs/error messages

2. **Use type examples throughout** - `?email=user@example.com&min_score=123`
   - ❌ When all params show examples, nothing stands out
   - ❌ Doesn't clearly indicate which parameter failed

3. **Use datatype.Example() placeholders** - `?email=abc&min_score=123`
   - ❌ Generic placeholders like 'abc' don't convey type information
   - ❌ Still doesn't guide attention to the problem

4. **Mixed approach with positional emphasis** ✅ **SELECTED**

## Decision

### Parameter Inclusion Rules

**We will ONLY include parameters that are:**

1. **Required by the endpoint definition**, OR
2. **Provided by the user in their request** (including the problematic parameter)

**We will EXCLUDE:**
- Optional parameters that the user did not provide
- Parameters not relevant to the current error

**Rationale**: Showing unused optional parameters creates noise and distracts from the actual problem. Users should only see parameters relevant to their specific request.

### Parameter Formatting Rules

**For error suggestion URLs, we will format parameters as:**

1. **Correct/non-problematic parameters** → UPPERCASE placeholders in braces
   - Format: `{PARAM_NAME}`
   - Examples: `{EMAIL}`, `{USER_ID}`, `{ACTIVE}`
   - Rationale: Low cognitive load, clearly indicates "fill this in"

2. **Failing/problematic parameter** → Actual example value from `Example()` method
   - Format: concrete value like `123`, `abc`, `user@example.com`
   - Rationale: Stands out visually, shows expected format
   - **Position: ALWAYS LAST** in the query string

3. **Visual hierarchy principle**: Position the problematic parameter LAST
   - Leverages natural reading pattern (eye lands on final element)
   - Creates visual contrast: placeholders → example value at end
   - Makes the error immediately obvious

### Examples

#### Example 1: Missing required email parameter (user provided min_score)
```
User requested: /api/users/search?min_score=50
Error: email is required

Suggestion: /api/users/search?min_score={MIN_SCORE}&email=user@example.com
                                                      ^^^^^^^^^^^^^^^^^^^^
                                                      Problem is here (LAST)

Note: Other optional params (limit, offset, etc.) NOT shown because user didn't provide them
```

#### Example 2: Invalid min_score parameter (email not required, user didn't provide it)
```
User requested: /api/users/search?min_score=invalid
Error: min_score must be an integer

Suggestion: /api/users/search?min_score=123
                               ^^^^^^^^^^^^^
                               Problem is here (LAST)

Note: email not shown because it's optional and user didn't provide it
```

#### Example 3: Multiple user-provided params, one failing
```
User requested: /api/posts?user_id=123&category=tech&limit=invalid
Error: limit must be a positive integer

Suggestion: /api/posts?user_id={USER_ID}&category={CATEGORY}&limit=50
                                                               ^^^^^^^
                                                               Problem is here (LAST)

Note: Shows only the 3 params the user provided
```

#### Example 4: Path parameter error with some query params
```
User requested: /api/users/invalid-id/posts?active=true
Error: user_id must be a valid integer

Suggestion: /api/users/{USER_ID}/posts?active={ACTIVE}
                       ^^^^^^^^^^
                       Problem is in path (can't move to LAST, but shown as example)

Note: active shown because user provided it; other optional query params omitted
```

#### Example 5: Optional parameter provided with invalid value
```
User requested: /api/search?category=tech&created_after=invalid-date
Error: created_after must be in YYYY-MM-DD format
Endpoint also has optional params: limit, offset, include_archived (user didn't provide)

Suggestion: /api/search?category={CATEGORY}&created_after=2024-01-15
                                             ^^^^^^^^^^^^^^^^^^^^^^
                                             Problem is here (LAST)

Note: limit, offset, include_archived NOT shown - user didn't use them
```

## Rationale

### Why Only Required + User-Provided Parameters?

1. **Reduces noise** - Don't overwhelm users with parameters they didn't use
2. **Contextual relevance** - Shows only what matters to THIS request
3. **Faster comprehension** - Shorter URLs are easier to scan
4. **Avoids confusion** - Users won't wonder "do I need to provide these optional params too?"
5. **Focused guidance** - Suggestion directly addresses their specific error

**Example of what to avoid:**
```
Bad: /api/search?category={CATEGORY}&limit={LIMIT}&offset={OFFSET}&
     include_archived={INCLUDE_ARCHIVED}&sort_by={SORT_BY}&
     created_after=2024-01-15

Good: /api/search?category={CATEGORY}&created_after=2024-01-15
```

### Why Placeholders for Correct Parameters?

1. **Reduced cognitive load** - `{EMAIL}` is instantly understood as "provide an email"
2. **No confusion with literals** - Clearly indicates a variable, not a fixed value
3. **Scannable** - Uppercase in braces stands out but doesn't dominate
4. **Consistent convention** - Matches URL templating patterns users already know

### Why Example Values for Problematic Parameters?

1. **Visual differentiation** - `123` looks different from `{MIN_SCORE}`
2. **Shows expected format** - `user@example.com` demonstrates what valid looks like
3. **Attention-grabbing** - Actual values "pop" when surrounded by placeholders
4. **Type clarity** - `2024-01-15` clearly shows it expects a date

### Why LAST Position?

1. **Reading pattern** - Western readers scan left-to-right, eye lands on rightmost element
2. **Visual hierarchy** - Final position = emphasis in query strings
3. **Progressive disclosure** - Context (correct params) → Problem (last param)
4. **Reduces search time** - User doesn't need to scan back and forth

## Implementation

### Parameter Classification and Filtering

When generating suggestion URLs:

```go
// Pseudocode
required_params = []
user_provided_params = []
problematic_params = []

// Step 1: Classify parameters
for each parameter in endpoint.AllParameters():
    if parameter.IsRequired():
        required_params.append(parameter)

    if user_provided(parameter):
        user_provided_params.append(parameter)

    if parameter.IsProblematic():
        problematic_params.append(parameter)

// Step 2: Determine which parameters to include
params_to_show = union(required_params, user_provided_params)

// Step 3: Format parameters
correct_params = []
failing_params = []

for each param in params_to_show:
    if param in problematic_params:
        value = param.Example()  // e.g., 123, user@example.com
        failing_params.append(param.Name() + "=" + value)
    else:
        value = "{" + param.Name().ToUpper() + "}"
        correct_params.append(param.Name() + "=" + value)

// Step 4: Build URL with correct params first, problematic last
url = path + "?" + join(correct_params, "&") + "&" + join(failing_params, "&")
```

### Example() Method Contract

Each parameter type must implement `Example()` to return representative values:

- `int`: `123`
- `string`: `abc`
- `email`: `user@example.com`
- `uuid`: `550e8400-e29b-41d4-a716-446655440000`
- `date`: `2024-01-15`
- `bool`: `true`
- `slug`: `my-slug`

These should be **simple, canonical examples** that clearly demonstrate the expected format without being verbose.

## Consequences

### Positive

1. **Improved UX** - Users immediately understand what to fix
2. **Focused suggestions** - Only shows relevant parameters
3. **Reduced support burden** - Clear suggestions = fewer "how do I fix this?" questions
4. **Consistent pattern** - Predictable format across all error responses
5. **Debuggable** - Easy to identify problematic parameters in logs
6. **Accessible** - Works well for screen readers (placeholders vs. values)
7. **Scalable** - Works well for endpoints with many optional parameters

### Negative

1. **Implementation complexity** - Requires parameter filtering and reordering logic
2. **Query string manipulation** - Must preserve semantics while changing order
3. **State tracking** - Must track which parameters user actually provided
4. **Multiple errors** - Need clear handling when multiple parameters fail
   - Solution: Show all failing params at end, separated from correct ones

### Risks

1. **Query parameter order sensitivity** - Some APIs/frameworks care about order
   - Mitigation: Document that suggestion URLs are illustrative, not literal
   - Note: Most REST APIs treat query params as order-independent

2. **Complex parameter relationships** - When params depend on each other
   - Mitigation: Show all related params together, still with problematic ones last

3. **Required param user didn't provide** - Shows in suggestion even though not in original request
   - This is CORRECT behavior: user needs to know about required params
   - Example: User calls `/api/search?category=tech`, endpoint requires `user_id`
   - Suggestion: `/api/search?category={CATEGORY}&user_id=123` ✓

## Multiple Failing Parameters

When multiple parameters fail:

```
User requested: /api/search?category=tech&email=bad&min_score=invalid
Errors: email invalid format, min_score must be integer

Suggestion: /api/search?category={CATEGORY}&email=user@example.com&min_score=123
                                            ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                            Both problems shown LAST
```

**Rule**: All failing parameters go after all correct parameters, maintaining the visual separation principle.

## Edge Cases

### All Parameters Fail
```
User requested: /api/search?email=bad&min_score=invalid
Both parameters fail

Suggestion: /api/search?email=user@example.com&min_score=123

Note: All params show example values, but still maintain logical order
```

### Only Required Parameter Fails (user provided nothing else)
```
User requested: /api/search
Error: email is required

Suggestion: /api/search?email=user@example.com

Note: No other params shown - user didn't provide any, and none others are required
```

## Future Considerations

1. **Interactive error messages** - Could highlight the problematic param in terminal/UI
2. **Severity levels** - Could use different formatting for warnings vs. errors
3. **Contextual help** - Could add inline comments like `&min_score=123 # must be positive`
4. **Localization** - Placeholder names might need translation in i18n scenarios
5. **Smart defaults** - Could show "commonly used" optional params even if not provided

## References

- RFC 9457: Problem Details for HTTP APIs
- [ADR-010: Error Handling with RFC 9457](./adr-010-error-handling-rfc9457.md)
- [ADR-016: Error Handling](./adr-016-error-handling.md)

---

**Decision makers**: Mike Schinkel, Claude Code
**Stakeholders**: API users, frontend developers, support team
