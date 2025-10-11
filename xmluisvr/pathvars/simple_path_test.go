package pathvars

import (
	"reflect"
	"testing"
)

func TestSimplePathWithoutParameters(t *testing.T) {
	// Test a simple path without any parameters - just the path part, not method+path
	template, err := ParseTemplate("/users")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	tests := []struct {
		path        string
		expectMatch bool
		expectErr   bool
	}{
		{path: "/users", expectMatch: true},
		{path: "/users/123", expectErr: true},
		{path: "/posts", expectErr: true},
		{path: "/users/", expectErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.path[1:], func(t *testing.T) {
			// Test exact match - should succeed
			vars, matched, _ := template.Match(tt.path, "")
			if tt.expectMatch && !matched {
				t.Errorf("Expected '%s' to match template '/users', but it didn't", tt.path)
			}
			if len(vars) != 0 {
				t.Errorf("Expected no variables for simple path, got %d: %v", len(vars), vars)
			}
		})
	}
}

func TestSimplePathWithoutParametersVerbose(t *testing.T) {
	// Test a simple path without any parameters
	template, err := ParseTemplate("/users")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	if template == nil {
		t.Fatalf("TEMPLATE IS NIL (SHOULD NOT HAPPEN, but Goland flag as potential nil reference")
	}

	//// Show what the regex looks like for debugging
	//t.Logf("Template regex: %v", template.regex)
	//if template.regex != nil {
	//	t.Logf("Regex pattern: %s", template.regex.String())
	//}

	tests := []struct {
		path     string
		expected bool
		name     string
		vm       ValuesMap
	}{
		{"/users", true, "exact match", nil},
		{"/posts", false, "different path", nil},
		{"/users/123", false, "path with extra segments", nil},
		{"/user", false, "similar but different path", nil},
		{"/users/", false, "path with trailing slash", nil},
	}

	for _, test := range tests {
		var valuesMap ValuesMap
		var matched bool
		valuesMap, matched, err = template.Match(test.path, "")
		//valuesMap, matched, err = template.Match(test.path, "")
		//t.Logf("Path: %s, Matched: %v, Vars: %v (%s)", test.path, matched, valuesMap, test.name)
		if err != nil {
			t.Errorf("Path error: %v", err)
		}

		if !reflect.DeepEqual(valuesMap, test.vm) {
			t.Errorf("Path expected: valuesMap=%v, got valuesMap=%v", test.vm, valuesMap)
		}

		if matched != test.expected {
			t.Errorf("Path expected: match=%v, got match=%v", test.expected, matched)
		}
	}
}
