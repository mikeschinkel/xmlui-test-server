package errparsr_test

//import (
//	"errors"
//	"fmt"
//	"testing"
//
//	"github.com/xmlui-org/xmlui-test-server/xmluisvr/errparsr"
//)
//
//var (
//	ErrBase   = errors.New("base error")
//	ErrOuter1 = errors.New("outer error 1")
//	ErrOuter2 = errors.New("outer error 2")
//	ErrLast   = errors.New("last error")
//)
//
//func TestUnwrapInsertThenJoin(t *testing.T) {
//	tests := []struct {
//		name               string
//		inner              error
//		outer              []error
//		fromLast           int
//		expectedOrder      []string
//		shouldContainBase  bool
//		shouldContainOuter bool
//		shouldContainLast  bool
//	}{
//		{
//			name: "insert outer before last error (fromLast=1)",
//			inner: doterr.NewErr(
//				ErrBase,
//				fmt.Errorf("detail1=value1"),
//				fmt.Errorf("detail2=value2"),
//				ErrLast,
//			),
//			outer: []error{
//				ErrOuter1,
//				ErrOuter2,
//				fmt.Errorf("extra=info"),
//			},
//			fromLast: 1,
//			expectedOrder: []string{
//				ErrBase.Error(),
//				"detail1=value1",
//				"detail2=value2",
//				ErrOuter1.Error(),
//				ErrOuter2.Error(),
//				"extra=info",
//				ErrLast.Error(),
//			},
//			shouldContainBase:  true,
//			shouldContainOuter: true,
//			shouldContainLast:  true,
//		},
//		{
//			name: "insert outer before last 2 errors (fromLast=2)",
//			inner: doterr.NewErr(
//				ErrBase,
//				fmt.Errorf("detail1=value1"),
//				fmt.Errorf("detail2=value2"),
//				ErrLast,
//			),
//			outer: []error{
//				ErrOuter1,
//				fmt.Errorf("extra=info"),
//			},
//			fromLast: 2,
//			expectedOrder: []string{
//				ErrBase.Error(),
//				"detail1=value1",
//				ErrOuter1.Error(),
//				"extra=info",
//				"detail2=value2",
//				ErrLast.Error(),
//			},
//			shouldContainBase:  true,
//			shouldContainOuter: true,
//			shouldContainLast:  true,
//		},
//		{
//			name: "insert at beginning (fromLast=all)",
//			inner: doterr.NewErr(
//				ErrBase,
//				ErrLast,
//			),
//			outer: []error{
//				ErrOuter1,
//			},
//			fromLast: 2,
//			expectedOrder: []string{
//				ErrOuter1.Error(),
//				ErrBase.Error(),
//				ErrLast.Error(),
//			},
//			shouldContainBase:  true,
//			shouldContainOuter: true,
//			shouldContainLast:  true,
//		},
//		{
//			name: "append at end (fromLast=0)",
//			inner: doterr.NewErr(
//				ErrBase,
//				ErrLast,
//			),
//			outer: []error{
//				ErrOuter1,
//			},
//			fromLast: 0,
//			expectedOrder: []string{
//				ErrBase.Error(),
//				ErrLast.Error(),
//				ErrOuter1.Error(),
//			},
//			shouldContainBase:  true,
//			shouldContainOuter: true,
//			shouldContainLast:  true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result := errparsr.UnwrapInsertThenJoin(tt.inner, tt.outer, tt.fromLast)
//
//			// Test 1: Verify errors.Is() works for all error types
//			if tt.shouldContainBase && !errors.Is(result, ErrBase) {
//				t.Errorf("Expected errors.Is(result, ErrBase) to be true")
//			}
//			if tt.shouldContainOuter && !errors.Is(result, ErrOuter1) {
//				t.Errorf("Expected errors.Is(result, ErrOuter1) to be true")
//			}
//			if tt.shouldContainLast && !errors.Is(result, ErrLast) {
//				t.Errorf("Expected errors.Is(result, ErrLast) to be true")
//			}
//
//			// Test 2: Verify error message order
//			resultStr := result.Error()
//			//t.Logf("Result error:\n%s", resultStr)
//
//			// Check order by finding index of each expected string
//			lastIndex := -1
//
//			for _, expected := range tt.expectedOrder {
//				// Find the expected string in result
//				idx := findErrorInString(resultStr, expected)
//				if idx == -1 {
//					t.Errorf("Expected error string not found: %q", expected)
//					continue
//				}
//				if idx < lastIndex {
//					t.Errorf("Error order incorrect: %q (index %d) should come after previous error (index %d)", expected, idx, lastIndex)
//				}
//				lastIndex = idx
//				//t.Logf("  [%d] Found %q at position %d", i, expected, idx)
//			}
//		})
//	}
//}
//
//// findErrorInString finds an error message within a multi-line joined error string
//func findErrorInString(haystack, needle string) int {
//	// errors.Join() separates errors with newlines
//	lines := splitErrorLines(haystack)
//	for i, line := range lines {
//		if line == needle {
//			return i
//		}
//	}
//	return -1
//}
//
//// splitErrorLines splits an error string by newlines
//func splitErrorLines(s string) []string {
//	var lines []string
//	current := ""
//	for _, ch := range s {
//		if ch == '\n' {
//			if current != "" {
//				lines = append(lines, current)
//				current = ""
//			}
//		} else {
//			current += string(ch)
//		}
//	}
//	if current != "" {
//		lines = append(lines, current)
//	}
//	return lines
//}
//
//func TestInsert(t *testing.T) {
//	tests := []struct {
//		name          string
//		inner         []error
//		outer         []error
//		fromLast      int
//		expectedOrder []string
//	}{
//		{
//			name: "basic insert before last",
//			inner: []error{
//				errors.New("first"),
//				errors.New("second"),
//				errors.New("third"),
//				errors.New("last"),
//			},
//			outer: []error{
//				errors.New("inserted"),
//			},
//			fromLast: 1,
//			expectedOrder: []string{
//				"first",
//				"second",
//				"third",
//				"inserted",
//				"last",
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result := errparsr.Insert(tt.inner, tt.outer, tt.fromLast)
//
//			if len(result) != len(tt.expectedOrder) {
//				t.Errorf("Expected %d errors, got %d", len(tt.expectedOrder), len(result))
//			}
//
//			for i, expected := range tt.expectedOrder {
//				if i >= len(result) {
//					t.Errorf("Missing error at index %d: %q", i, expected)
//					continue
//				}
//				if result[i].Error() != expected {
//					t.Errorf("At index %d: expected %q, got %q", i, expected, result[i].Error())
//				}
//			}
//		})
//	}
//}
