package jsonutil_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/jsonutil"
)

func TestStreamingExtractValue(t *testing.T) {
	type tc struct {
		name         string
		raw          string
		query        dbqvars.Parameter
		want         any
		wantErrIsAny []error // at least one of these must match via errors.Is
		wantErrIsAll []error // all of these must match via errors.Is
	}

	tests := []tc{
		{
			name:  "simple object key",
			raw:   `{"foo":"bar"}`,
			query: dbqvars.Parameter("foo"),
			want:  "bar",
		},
		{
			name:  "nested object key",
			raw:   `{"a":{"b":{"c":123}}}`,
			query: dbqvars.Parameter("a.b.c"),
			want:  float64(123), // json.Unmarshal numbers → float64 in interface{}
		},
		{
			name:  "array index",
			raw:   `{"xs":[10,20,30]}`,
			query: dbqvars.Parameter("xs.1"),
			want:  float64(20),
		},
		{
			name:  "array of objects then key",
			raw:   `{"foo":[{"bar":"baz"}]}`,
			query: dbqvars.Parameter("foo.0.bar"),
			want:  "baz",
		},
		{
			name:  "array index out of range",
			raw:   `{"xs":[10,20,30]}`,
			query: dbqvars.Parameter("xs.3"),
			wantErrIsAny: []error{
				jsonutil.ErrJSONIndexOutOfRange,
			},
		},
		{
			name:  "missing object key",
			raw:   `{"obj":{"have":1}}`,
			query: dbqvars.Parameter("obj.missing"),
			wantErrIsAny: []error{
				jsonutil.ErrJSONPathSegmentNotFound,
			},
		},
		{
			name:  "expects array but found object",
			raw:   `{"obj":{"k":1}}`,
			query: dbqvars.Parameter("obj.0"),
			wantErrIsAny: []error{
				jsonutil.ErrJSONPathExpectedArrayAtSegment,
			},
		},
		{
			name:  "expects object but found array",
			raw:   `{"xs":[{"k":1}]}`,
			query: dbqvars.Parameter("xs.k"),
			wantErrIsAny: []error{
				jsonutil.ErrJSONPathExpectedObjectAtSegment,
			},
		},
		{
			name:  "empty segment in path",
			raw:   `{"a":{"b":1}}`,
			query: dbqvars.Parameter("a..b"),
			wantErrIsAny: []error{
				jsonutil.ErrJSONPathContainsEmptySegment,
			},
		},
		{
			name:  "negative index",
			raw:   `{"xs":[0,1]}`,
			query: dbqvars.Parameter("xs.-1"),
			wantErrIsAny: []error{
				jsonutil.ErrJSONIndexOutOfRange,
			},
		},
		{
			name:  "empty body",
			raw:   ``,
			query: dbqvars.Parameter("foo"),
			wantErrIsAll: []error{
				jsonutil.ErrJSONPathTraversalFailed,
				jsonutil.ErrJSONBodyCannotBeEmpty,
			},
		},
		{
			name:  "empty SQL parameter name",
			raw:   `{"foo":"bar"}`,
			query: dbqvars.Parameter(""),
			wantErrIsAny: []error{
				jsonutil.ErrJSONValueSelectorCannotBeEmpty,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonutil.ExtractValueFromBytes([]byte(tt.raw), common.Selector(tt.query))

			// Error expectations
			if len(tt.wantErrIsAny) > 0 || len(tt.wantErrIsAll) > 0 {
				if err == nil {
					t.Fatalf("ExtractValueFromBytes() expected an error, got nil (value=%#v)", got)
				}
				// any-of
				if len(tt.wantErrIsAny) > 0 {
					okAny := false
					for _, we := range tt.wantErrIsAny {
						if errors.Is(err, we) {
							okAny = true
							break
						}
					}
					if !okAny {
						t.Fatalf("ExtractValueFromBytes() error %v did not match any of expected %v", err, tt.wantErrIsAny)
					}
				}
				// all-of
				for _, we := range tt.wantErrIsAll {
					if !errors.Is(err, we) {
						t.Fatalf("ExtractValueFromBytes() error %v is not errors.Is(...) to %v", err, we)
					}
				}
				return
			}

			// Success expectations
			if err != nil {
				t.Fatalf("ExtractValueFromBytes() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ExtractValueFromBytes() value mismatch:\n  got:  %#v (%T)\n  want: %#v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestStreamingExtractValue_UnmarshalErrorIsWrapped(t *testing.T) {
	// invalid JSON
	_, err := jsonutil.ExtractValueFromBytes([]byte(`{"unterminated": 1`), "foo")
	if err == nil {
		t.Fatalf("ExtractValueFromBytes() expected error for invalid JSON, got nil")
	}
	// With streaming, JSON syntax errors are wrapped in ErrJSONTokenReadFailed
	if !errors.Is(err, jsonutil.ErrJSONTokenReadFailed) {
		t.Fatalf("ExtractValueFromBytes() error %v not errors.Is(..., ErrJSONTokenReadFailed)", err)
	}
}

func TestStreamingExtractValue_TypeTransparence(t *testing.T) {
	// Ensure numbers come back as float64 and booleans/strings/objects are preserved.
	raw := `{"n": 1, "b": true, "s": "x", "o": {"k": "v"}, "a": [1,2]}`

	check := func(path string, want any) {
		got, err := jsonutil.ExtractValueFromBytes([]byte(raw), common.Selector(path))
		if err != nil {
			t.Fatalf("ExtractValueFromBytes(%q) unexpected error: %v", path, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ExtractValueFromBytes(%q) got %#v (%T), want %#v (%T)", path, got, got, want, want)
		}
	}

	check("n", float64(1))
	check("b", true)
	check("s", "x")
	check("o.k", "v")
	check("a.0", float64(1))
}

func TestStreamingExtractValue_WithReader(t *testing.T) {
	jsonData := `{"user": {"profile": {"name": "Alice", "age": 30}}, "scores": [100, 85, 92]}`

	tests := []struct {
		name string
		path string
		want any
	}{
		{"nested object", "user.profile.name", "Alice"},
		{"nested number", "user.profile.age", float64(30)},
		{"array element", "scores.1", float64(85)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(jsonData)
			got, err := jsonutil.ExtractValueFromReader(reader, common.Selector(tt.path))
			if err != nil {
				t.Fatalf("ExtractValueFromReader() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ExtractValueFromReader() got %#v (%T), want %#v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestStreamingExtractValue_ErrorContext(t *testing.T) {

	// Test that error messages contain helpful context
	jsonData := `{"users": [{"name": "Alice"}, {"name": "Bob"}], "settings": {"theme": "dark"}}`

	_, err := jsonutil.ExtractValueFromBytes([]byte(jsonData), "users.5.name")
	if err == nil {
		t.Fatal("Expected error for out of range array index")
	}

	// Check that error contains context information
	errStr := err.Error()
	if !strings.Contains(errStr, "target_index=5") {
		t.Errorf("Error should contain target index: %v", err)
	}
	if !strings.Contains(errStr, "array_length=2") {
		t.Errorf("Error should contain array length: %v", err)
	}
}

func TestExtractValuesFromBytes_MultipleSelectors(t *testing.T) {
	jsonData := `{
		"user": {"name": "Alice", "age": 30},
		"scores": [100, 85, 92],
		"settings": {"theme": "dark", "lang": "en"}
	}`

	tests := []struct {
		name       string
		selectors  []common.Selector
		wantValues []any
		wantFound  []common.Selector
		wantErr    bool
	}{
		{
			name: "multiple valid selectors",
			selectors: []common.Selector{
				"user.name",
				"user.age",
				"scores.1",
				"settings.theme",
			},
			wantValues: []any{"Alice", float64(30), float64(85), "dark"},
			wantFound:  []common.Selector{"user.name", "user.age", "scores.1", "settings.theme"},
			wantErr:    false,
		},
		{
			name: "mixed valid and invalid selectors",
			selectors: []common.Selector{
				"user.name",    // valid
				"user.missing", // invalid
				"scores.0",     // valid
				"scores.10",    // invalid - out of range
			},
			wantValues: []any{"Alice", nil, float64(100), nil},
			wantFound:  []common.Selector{"user.name", "scores.0"},
			wantErr:    true,
		},
		{
			name: "all invalid selectors",
			selectors: []common.Selector{
				"missing.key",
				"user.nonexistent",
				"scores.999",
			},
			wantValues: []any{nil, nil, nil},
			wantFound:  []common.Selector{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, found, err := jsonutil.ExtractValuesFromBytes([]byte(jsonData), tt.selectors)

			// Check error expectation
			if tt.wantErr && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check found selectors
			if !reflect.DeepEqual(found, tt.wantFound) {
				t.Errorf("Found selectors mismatch:\n  got:  %v\n  want: %v", found, tt.wantFound)
			}

			// Check values for found selectors
			for _, selector := range tt.wantFound {
				selectorIndex := -1
				for j, s := range tt.selectors {
					if s == selector {
						selectorIndex = j
						break
					}
				}
				if selectorIndex == -1 {
					t.Fatalf("Found selector %s not in original selectors", selector)
				}

				expectedValue := tt.wantValues[selectorIndex]
				actualValue := values[selectorIndex]
				if !reflect.DeepEqual(actualValue, expectedValue) {
					t.Errorf("Value mismatch for selector %s:\n  got:  %#v (%T)\n  want: %#v (%T)",
						selector, actualValue, actualValue, expectedValue, expectedValue)
				}
			}
		})
	}
}

func TestExtractValuesFromReader_MultipleSelectors(t *testing.T) {
	jsonData := `{"a": 1, "b": {"c": 2}, "d": [3, 4, 5]}`

	selectors := []common.Selector{"a", "b.c", "d.2"}

	reader := strings.NewReader(jsonData)
	values, found, err := jsonutil.ExtractValuesFromReader(reader, selectors)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedValues := []any{float64(1), float64(2), float64(5)}
	expectedFound := []common.Selector{"a", "b.c", "d.2"}

	if !reflect.DeepEqual(found, expectedFound) {
		t.Errorf("Found selectors mismatch:\n  got:  %v\n  want: %v", found, expectedFound)
	}

	for i, expected := range expectedValues {
		if !reflect.DeepEqual(values[i], expected) {
			t.Errorf("Value %d mismatch:\n  got:  %#v (%T)\n  want: %#v (%T)",
				i, values[i], values[i], expected, expected)
		}
	}
}

func TestExtractValuesFromBytes_ErrorCollection(t *testing.T) {
	jsonData := `{"valid": "value"}`

	// Multiple invalid selectors to test error collection
	selectors := []common.Selector{
		"missing1",
		"missing2",
		"valid", // This one should succeed
		"missing3",
	}

	values, found, err := jsonutil.ExtractValuesFromBytes([]byte(jsonData), selectors)

	// Should have error for the missing selectors
	if err == nil {
		t.Fatal("Expected error for missing selectors")
	}

	// Should find the valid selector
	expectedFound := []common.Selector{"valid"}
	if !reflect.DeepEqual(found, expectedFound) {
		t.Errorf("Found selectors mismatch:\n  got:  %v\n  want: %v", found, expectedFound)
	}

	// Should have the valid value
	if values[2] != "value" {
		t.Errorf("Expected valid value 'value', got %v", values[2])
	}

	// Error should contain information about all missing selectors
	errStr := err.Error()
	if !strings.Contains(errStr, "missing1") {
		t.Errorf("Error should mention missing1: %v", err)
	}
	if !strings.Contains(errStr, "missing2") {
		t.Errorf("Error should mention missing2: %v", err)
	}
	if !strings.Contains(errStr, "missing3") {
		t.Errorf("Error should mention missing3: %v", err)
	}
}
