package dbqvars

import (
	"unicode"
)

var _ ParsedQuery = (*ParsedSQL)(nil)

type ParsedSQL struct {
	SQL        SQLQuery
	parameters []Parameter // ordered by first appearance, deduped by Name
}

func NewParsedSQL(SQL SQLQuery, parameters []Parameter) ParsedSQL {
	if parameters == nil {
		parameters = make([]Parameter, 0)
	}
	return ParsedSQL{
		SQL:        SQL,
		parameters: parameters,
	}
}

//func (ps ParsedSQL) GetValues(paramsMap map[Identifier]any, bodyJSON []byte) (values []any, err error) {
//	return QueryTokens(ps.Parameters).GetValues(paramsMap, bodyJSON)
//}

func (ps ParsedSQL) QueryString() QueryString {
	return QueryString(ps.SQL)
}

func (ps ParsedSQL) Parameters() (names Parameters) {
	return ps.parameters
}

type ParseSQLArgs struct{}

// ParseSQL finds {name} placeholders OUTSIDE of strings/identifiers/comments,
// rewrites them via FormatParamFunc, and returns the rewritten SQL & ordered tokens.
// FormatParamFunc examples:
//
//	Postgres: func(i int) string { return fmt.Sprintf("$%d", i) }
//	MySQL/SQLite: func(int) string { return "?" }
//	SQL Server: func(i int) string { return fmt.Sprintf("@p%d", i) }
func ParseSQL(sqlText SQLQuery, formatFunc FormatParamFunc) (ps ParsedSQL, err error) {
	var state parseState

	if formatFunc == nil {
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
			err = state.consumePlaceholder(formatFunc)
			if err != nil {
				goto end
			}
			continue
		}

		state.i++
	}

	if len(state.edits) == 0 {
		ps = NewParsedSQL(
			SQLQuery(state.src),
			state.tokens.Parameters(),
		)
		goto end
	}

	ps = NewParsedSQL(
		state.buildSQL(),
		state.orderedTokens().Parameters(),
	)

end:
	return ps, err
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
				if !readDigits(s, &i) {
					goto end
				}
			}
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

func readDigits(s string, i *int) (ok bool) {
	start := *i

	if *i >= len(s) {
		goto end
	}

	for *i < len(s) {
		if s[*i] < '0' || s[*i] > '9' {
			break
		}
		*i++
	}

	if *i == start {
		goto end
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
