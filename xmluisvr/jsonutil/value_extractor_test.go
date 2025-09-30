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
			query: dbqvars.NewParameter("foo", 1),
			want:  "bar",
		},
		{
			name:  "nested object key",
			raw:   `{"a":{"b":{"c":123}}}`,
			query: dbqvars.NewParameter("a.b.c", 1),
			want:  float64(123), // json.Unmarshal numbers → float64 in interface{}
		},
		{
			name:  "array index",
			raw:   `{"xs":[10,20,30]}`,
			query: dbqvars.NewParameter("xs.1", 1),
			want:  float64(20),
		},
		{
			name:  "array of objects then key",
			raw:   `{"foo":[{"bar":"baz"}]}`,
			query: dbqvars.NewParameter("foo.0.bar", 1),
			want:  "baz",
		},
		{
			name:  "array index out of range",
			raw:   `{"xs":[10,20,30]}`,
			query: dbqvars.NewParameter("xs.3", 1),
			wantErrIsAny: []error{
				jsonutil.ErrJSONIndexOutOfRange,
			},
		},
		{
			name:  "missing object key",
			raw:   `{"obj":{"have":1}}`,
			query: dbqvars.NewParameter("obj.missing", 1),
			wantErrIsAny: []error{
				jsonutil.ErrJSONPathSegmentNotFound,
			},
		},
		{
			name:  "expects array but found object",
			raw:   `{"obj":{"k":1}}`,
			query: dbqvars.NewParameter("obj.0", 1),
			wantErrIsAny: []error{
				jsonutil.ErrJSONPathExpectedArrayAtSegment,
			},
		},
		{
			name:  "expects object but found array",
			raw:   `{"xs":[{"k":1}]}`,
			query: dbqvars.NewParameter("xs.k", 1),
			wantErrIsAny: []error{
				jsonutil.ErrJSONPathExpectedObjectAtSegment,
			},
		},
		{
			name:  "empty segment in path",
			raw:   `{"a":{"b":1}}`,
			query: dbqvars.NewParameter("a..b", 1),
			wantErrIsAny: []error{
				jsonutil.ErrJSONPathContainsEmptySegment,
			},
		},
		{
			name:  "negative index",
			raw:   `{"xs":[0,1]}`,
			query: dbqvars.NewParameter("xs.-1", 1),
			wantErrIsAny: []error{
				jsonutil.ErrJSONIndexOutOfRange,
			},
		},
		{
			name:  "empty body",
			raw:   ``,
			query: dbqvars.NewParameter("foo", 1),
			wantErrIsAll: []error{
				jsonutil.ErrJSONPathTraversalFailed,
				jsonutil.ErrJSONBodyCannotBeEmpty,
			},
		},
		{
			name:  "empty SQL parameter name",
			raw:   `{"foo":"bar"}`,
			query: dbqvars.NewParameter("", 1),
			wantErrIsAny: []error{
				jsonutil.ErrJSONValueSelectorCannotBeEmpty,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonutil.ExtractValueFromBytes([]byte(tt.raw), tt.query.Name)

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
		name          string
		selectors     []common.Selector
		wantValuesMap jsonutil.ValuesMap
		wantNotFound  []common.Selector
		wantErr       bool
	}{
		{
			name: "multiple valid selectors",
			selectors: []common.Selector{
				"user.name",
				"user.age",
				"scores.1",
				"settings.theme",
			},
			wantValuesMap: jsonutil.ValuesMap{
				"user.name":      "Alice",
				"user.age":       float64(30),
				"scores.1":       float64(85),
				"settings.theme": "dark",
			},
			wantNotFound: []common.Selector{},
			wantErr:      false,
		},
		{
			name: "mixed valid and invalid selectors",
			selectors: []common.Selector{
				"user.name",    // valid
				"user.missing", // invalid
				"scores.0",     // valid
				"scores.10",    // invalid - out of range
			},
			wantValuesMap: jsonutil.ValuesMap{
				"user.name": "Alice",
				"scores.0":  float64(100),
			},
			wantNotFound: []common.Selector{"user.missing", "scores.10"},
			wantErr:      true,
		},
		{
			name: "all invalid selectors",
			selectors: []common.Selector{
				"missing.key",
				"user.nonexistent",
				"scores.999",
			},
			wantValuesMap: jsonutil.ValuesMap{},
			wantNotFound:  []common.Selector{"missing.key", "user.nonexistent", "scores.999"},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valuesMap, notFound, err := jsonutil.ExtractValuesFromBytes([]byte(jsonData), tt.selectors)

			// Check error expectation
			if tt.wantErr && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check notFound selectors
			if !reflect.DeepEqual(notFound, tt.wantNotFound) {
				t.Errorf("NotFound selectors mismatch:\n  got:  %v\n  want: %v", notFound, tt.wantNotFound)
			}

			// Check valuesMap
			if !reflect.DeepEqual(valuesMap, tt.wantValuesMap) {
				t.Errorf("ValuesMap mismatch:\n  got:  %#v\n  want: %#v", valuesMap, tt.wantValuesMap)
			}
		})
	}
}

func TestExtractValuesFromReader_MultipleSelectors(t *testing.T) {
	jsonData := `{"a": 1, "b": {"c": 2}, "d": [3, 4, 5]}`

	selectors := []common.Selector{"a", "b.c", "d.2"}

	reader := strings.NewReader(jsonData)
	valuesMap, notFound, err := jsonutil.ExtractValuesFromReader(reader, selectors)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedValuesMap := jsonutil.ValuesMap{
		"a":   float64(1),
		"b.c": float64(2),
		"d.2": float64(5),
	}
	expectedNotFound := []common.Selector{}

	if !reflect.DeepEqual(notFound, expectedNotFound) {
		t.Errorf("NotFound selectors mismatch:\n  got:  %v\n  want: %v", notFound, expectedNotFound)
	}

	if !reflect.DeepEqual(valuesMap, expectedValuesMap) {
		t.Errorf("ValuesMap mismatch:\n  got:  %#v\n  want: %#v", valuesMap, expectedValuesMap)
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

	valuesMap, notFound, err := jsonutil.ExtractValuesFromBytes([]byte(jsonData), selectors)

	// Should have error for the missing selectors
	if err == nil {
		t.Fatal("Expected error for missing selectors")
	}

	// Should have the not found selectors
	expectedNotFound := []common.Selector{"missing1", "missing2", "missing3"}
	if !reflect.DeepEqual(notFound, expectedNotFound) {
		t.Errorf("NotFound selectors mismatch:\n  got:  %v\n  want: %v", notFound, expectedNotFound)
	}

	// Should have the valid value in the map
	if valuesMap["valid"] != "value" {
		t.Errorf("Expected valid value 'value', got %v", valuesMap["valid"])
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
