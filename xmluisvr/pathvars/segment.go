package pathvars

// Segment represents a part of the path template
type Segment string

// IsLiteral returns true if this segment is a literal string
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

// IsParameter returns true if this segment is a parameter
func (s Segment) IsParameter() (is bool) {
	return !s.IsLiteral()
}
