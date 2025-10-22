package errparsr

//
//import (
//	"testing"
//)
//
//func ShowTree(t *testing.T, err error, level ...int) {
//	errs, unwrapped := UnwrapJoinErrors(err)
//	if unwrapped {
//		t.Logf("%#v", errs)
//		goto end
//	}
//	for _, ue := range errs {
//		if !IsJoinError(ue) {
//			t.Logf("%#v", errs)
//			continue
//		}
//		t.Logf("%#v", errs)
//		n := 0
//		if len(level) > 0 {
//			n = level[0]
//		}
//		ShowTree(t, ue, n)
//	}
//end:
//	return
//}
