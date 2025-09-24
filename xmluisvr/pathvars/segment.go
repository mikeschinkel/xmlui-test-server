// Package pathvars/segment defines the Segment type which represents
// individual parts of a path template. Segments can be either literal
// strings or parameter placeholders enclosed in braces.
package pathvars

// Segment represents a part of the path template, either a literal string
// or a parameter placeholder like {id:int}. Segments are used during
// template parsing and regex generation.
type Segment string

// IsLiteral returns true if this segment is a literal string (not a parameter).
// Literal segments are used as-is in URL paths without any substitution.
func (s Segment) IsLiteral() (isLit bool) {
	if len(s) == 0 {
		goto end
	}
	if s[0] == '{' {
		goto end
	}
	isLit = true
end:
	return isLit
}

// IsParameter returns true if this segment is a parameter placeholder.
// Parameter segments are enclosed in braces and contain parameter definitions.
func (s Segment) IsParameter() (is bool) {
	return !s.IsLiteral()
}
