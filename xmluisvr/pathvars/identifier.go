package pathvars

import (
	"errors"
	"fmt"
	"regexp"
)

var identifierRegex = regexp.MustCompile(`^[a-zA-Z_]\w*$`)

func isLetter(c byte) (is bool) {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
func ParseIdentifier(s string) (id Identifier, err error) {
	var match string
	if s == "" {
		err = ErrIdentifierCannotBeEmpty
		goto end
	}
	match = identifierRegex.FindString(s)
	if match != "" {
		id = Identifier(match)
		goto end
	}
	if s[0] != '_' && !isLetter(s[0]) {
		err = errors.Join(
			ErrInvalidIdentifier,
			ErrMustBeginWithLetterOrUnderscore,
			fmt.Errorf("value=%s", s),
		)
		goto end
	}
	err = errors.Join(
		ErrInvalidNameSpec,
		ErrMustOnlyContainLettersDigitsAndOrUnderscores,
		fmt.Errorf("value=%s", s),
	)
	id = Identifier(s)
end:
	return id, err
}

var leadingIdentifierRegex = regexp.MustCompile(`^(\w+)(\W*)`)

func ParseLeadingIdentifier(s string) (id Identifier, err error) {
	matches := leadingIdentifierRegex.FindStringSubmatch(s)
	if matches == nil {
		err = errors.Join(
			ErrInvalidIdentifier,
			ErrMustBeginWithLetterOrUnderscore,
			fmt.Errorf("value=%s", s),
		)
		goto end
	}
	id = Identifier(matches[1])
end:
	return id, err
}

var (
	ErrInvalidIdentifier                            = errors.New("invalid identifier")
	ErrMustBeginWithLetterOrUnderscore              = errors.New("must begin with letter or underscore")
	ErrMustOnlyContainLettersDigitsAndOrUnderscores = errors.New("must only contain letters, digits, and/or underscores")
	ErrIdentifierCannotBeEmpty                      = errors.New("identifier can not be empty")
)
