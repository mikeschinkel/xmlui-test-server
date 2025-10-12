package errparsr

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
)

// IsCustomError reports whether err’s concrete type comes from a non-stdlib module.
func IsCustomError(err error) (isCustom bool) {
	var t reflect.Type

	if err == nil {
		goto end
	}
	t = reflect.TypeOf(err)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	isCustom = strings.Contains(t.PkgPath(), ".") // non-stdlib packages have a dot
end:
	return isCustom
}

var joinErrorArchetype = reflect.TypeOf(errors.Join(errors.New("x"), errors.New("y")))

func IsJoinError(err error) (isJoin bool) {
	var jl interface{ Unwrap() []error }
	if err == nil {
		goto end
	}
	// direct match
	if reflect.TypeOf(err) == joinErrorArchetype {
		isJoin = true
		goto end
	}
	if !errors.As(err, &jl) {
		goto end
	}
	// see through wrapping
	isJoin = reflect.TypeOf(jl) == joinErrorArchetype
end:
	return isJoin
}

var captureDetailsRegex = regexp.MustCompile(`^(\w+)=(.*)$`)

func UnwrapJoinErrors(err error) []error {
	u, ok := err.(interface {
		Unwrap() []error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

func UnwrapAllJoinErrors(err error) (all []error) {
	for _, ue := range UnwrapJoinErrors(err) {
		if !IsJoinError(ue) {
			all = append(all, ue)
			continue
		}
		all = append(all, UnwrapAllJoinErrors(ue)...)
	}
	return all
}
