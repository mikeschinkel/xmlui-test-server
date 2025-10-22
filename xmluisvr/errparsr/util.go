package errparsr

//
//import (
//	"errors"
//	"fmt"
//	"reflect"
//	"regexp"
//	"strings"
//
//	"github.com/xmlui-org/xmlui-test-server/xmluisvr/doterr"
//)
//
//// IsCustomError reports whether err’s concrete type comes from a non-stdlib module.
//func IsCustomError(err error) (isCustom bool) {
//	var t reflect.Type
//
//	if err == nil {
//		goto end
//	}
//	t = reflect.TypeOf(err)
//	if t.Kind() == reflect.Ptr {
//		t = t.Elem()
//	}
//	isCustom = strings.Contains(t.PkgPath(), ".") // non-stdlib packages have a dot
//end:
//	return isCustom
//}
//
//var joinErrorArchetype = reflect.TypeOf(doterr.NewErr(errors.New("x"), errors.New("y")))
//
//func IsJoinError(err error) (isJoin bool) {
//	var jl interface{ Unwrap() []error }
//	if err == nil {
//		goto end
//	}
//	// direct match
//	if reflect.TypeOf(err) == joinErrorArchetype {
//		isJoin = true
//		goto end
//	}
//	if !errors.As(err, &jl) {
//		goto end
//	}
//	// see through wrapping
//	isJoin = reflect.TypeOf(jl) == joinErrorArchetype
//end:
//	return isJoin
//}
//
//var captureDetailsRegex = regexp.MustCompile(`^(\w+)=(.*)$`)
//
//func UnwrapJoinErrors(err error) (errs []error, unwrapped bool) {
//	u, ok := err.(interface {
//		Unwrap() []error
//	})
//	if !ok {
//		errs = []error{err}
//		goto end
//	}
//	unwrapped = true
//	errs = u.Unwrap()
//end:
//	return errs, unwrapped
//}
//
//func UnwrapAllJoinErrors(err error) (all []error) {
//	errs, unwrapped := UnwrapJoinErrors(err)
//	if !unwrapped {
//		all = errs
//		goto end
//	}
//	for _, ue := range errs {
//		if !IsJoinError(ue) {
//			all = append(all, ue)
//			continue
//		}
//		all = append(all, UnwrapAllJoinErrors(ue)...)
//	}
//end:
//	return all
//}
//
//type unwrapperN interface{ Unwrap() []error }
//type unwrapper1 interface{ Unwrap() error }
//
//func UnwrapInsertThenJoin(inner error, outer []error, fromLast int) (err error) {
//	return errors.Join(Insert(UnwrapAllJoinErrors(inner), outer, fromLast)...)
//}
//func Insert(inner []error, outer []error, fromLast int) (errs []error) {
//	pos := len(inner) - fromLast
//	if pos < 0 {
//		panic(fmt.Sprintf("Invalid 'fromLast' parameter %d passed to errparsr.Insert(); cannot be greater than length (%d) of 'inner' parameter", fromLast, len(inner)))
//	}
//	errs = append(errs, inner[:pos]...) // First part of inner errors
//	errs = append(errs, outer...)       // Insert outer errors
//	errs = append(errs, inner[pos:]...) // Last part of inner errors
//	return errs
//}
//
//func isEmpty(s string) (empty bool) {
//	empty = true
//	if s == "" {
//		goto end
//	}
//	for c := range s {
//		switch c {
//		case ' ', '\t', '\r', '\n':
//			continue
//		default:
//			empty = false
//			goto end
//		}
//	}
//end:
//	return empty
//}
