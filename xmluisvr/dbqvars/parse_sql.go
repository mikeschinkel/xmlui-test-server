package dbqvars

import (
	"unicode"
)

type SQLParam struct {
	Name  string // logical name: e.g., "path.accountId" or "body.items[0].id"
	Index int    // assigned parameter index (1-based)
	Start int    // byte offset start in original SQL
	End   int    // byte offset end (exclusive)
	Raw   string // full token, e.g. "{user.id}"
}

type ParsedSQL struct {
	SQL    string
	Params []SQLParam // ordered by first appearance, deduped by Name
}

type ParseSQLArgs struct {
	// Required: render a bind placeholder for the given 1-based position.
	// Examples: Postgres: func(i int) string { return fmt.Sprintf("$%d", i) }
	//           MySQL/SQLite: func(int) string { return "?" }
	//           SQL Server: func(i int) string { return fmt.Sprintf("@p%d", i) }
	FormatParamFunc FormatParamFunc
}

// ParseSQL finds {name} placeholders OUTSIDE of strings/identifiers/comments,
// rewrites them via FormatParamFunc, and returns the rewritten SQL & ordered params.
func ParseSQL(sqlText string, args ParseSQLArgs) (sql ParsedSQL, err error) {
	var state parseState
	var ordered []SQLParam

	if args.FormatParamFunc == nil {
		err = ErrFormatParamFuncRequired
		goto end
	}

	state = newParseState(sqlText)

	for state.i < state.n {
		c := state.src[state.i]

		switch c {
		case '-':
			if state.peek(1) == '-' {
				state.i += 2
				state.consumeDashDash()
				continue
			}
		case '#':
			state.consumeHashComment()
			continue
		case '/':
			if state.peek(1) == '*' {
				state.consumeBlockComment()
				continue
			}
		case '\'':
			state.consumeSingleQuoted()
			continue
		case '"':
			state.consumeDoubleQuoted()
			continue
		case '`':
			state.consumeBacktick()
			continue
		case '[':
			state.consumeBracketIdent()
			continue
		case '$':
			state.consumeDollarQuoted()
			continue
		case 'q', 'Q':
			state.consumeOracleQ()
			continue
		case '{':
			err = state.consumePlaceholder(args.FormatParamFunc)
			if err != nil {
				goto end
			}
			continue
		}

		state.i++
	}

	if len(state.edits) == 0 {
		sql = ParsedSQL{SQL: state.src, Params: state.orderParams()}
		goto end
	}

	ordered = state.orderParams()
	sql = ParsedSQL{SQL: state.buildSQL(), Params: ordered}

end:
	return sql, err
}

func isValidName(s string) (is bool) {
	var i int
	if s == "" {
		goto end
	}
	i = 0
	if !readIdent(s, &i) {
		goto end
	}
	for i < len(s) {
		switch s[i] {
		case '.':
			i++
			if !readIdent(s, &i) {
				goto end
			}
		case '[':
			i++
			start := i
			for {
				if i >= len(s) {
					break
				}
				if s[i] < '0' {
					break
				}
				if s[i] > '9' {
					break
				}
				i++
			}
			if i == start {
				goto end
			}
			if i >= len(s) {
				goto end
			}
			if s[i] != ']' {
				goto end
			}
			i++
		default:
			goto end
		}
	}
	is = true
end:
	return is
}

func readIdent(s string, i *int) (ok bool) {
	var r rune
	var w int

	if *i >= len(s) {
		goto end
	}

	r, w = utf8At(s, *i)
	if !isLetterOrUnderscore(r) {
		goto end
	}

	*i += w
	for *i < len(s) {
		r, w := utf8At(s, *i)
		if isLetterDigitOrUnderscore(r) {
			*i += w
			continue
		}
		break
	}

	ok = true

end:
	return ok
}

// utf8At returns the rune at byte offset i and its width.
func utf8At(s string, i int) (r rune, w int) {
	var rr []rune
	// Fast path: ASCII
	b := s[i]
	if b < 0x80 {
		r, w = rune(b), 1
		goto end
	}
	// Minimal safe decode for a single rune
	rr = []rune(s[i:])
	r, w = rr[0], len(string(rr[:1]))
end:
	return r, w
}

func isLetterOrUnderscore(r rune) (is bool) {
	if r == '_' {
		is = true
		goto end
	}
	if unicode.IsLetter(r) {
		is = true
		goto end
	}
end:
	return is
}

func isLetterDigitOrUnderscore(r rune) (is bool) {
	if r == '_' {
		is = true
		goto end
	}
	if unicode.IsLetter(r) {
		is = true
		goto end
	}
	if unicode.IsDigit(r) {
		is = true
		goto end
	}
end:
	return is
}
