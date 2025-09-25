package pathvars

import (
	"testing"
)

func TestSimplePathWithoutParameters(t *testing.T) {
	// Test a simple path without any parameters - just the path part, not method+path
	template, err := ParseTemplate("/users")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Test exact match - should succeed
	vars, matched := template.Match("/users", "")
	if !matched {
		t.Errorf("Expected /users to match template '/users', but it didn't")
	}
	if len(vars) != 0 {
		t.Errorf("Expected no variables for simple path, got %d: %v", len(vars), vars)
	}

	// Test non-matching path - should fail
	_, matched2 := template.Match("/posts", "")
	if matched2 {
		t.Errorf("Expected /posts to NOT match template '/users', but it did")
	}

	// Test path with extra segments - should fail
	_, matched3 := template.Match("/users/123", "")
	if matched3 {
		t.Errorf("Expected /users/123 to NOT match template '/users', but it did")
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

	// Show what the regex looks like for debugging
	t.Logf("Template regex: %v", template.regex)
	if template.regex != nil {
		t.Logf("Regex pattern: %s", template.regex.String())
	}

	tests := []struct {
		path     string
		expected bool
		name     string
	}{
		{"/users", true, "exact match"},
		{"/posts", false, "different path"},
		{"/users/123", false, "path with extra segments"},
		{"/user", false, "similar but different path"},
		{"/users/", false, "path with trailing slash"},
	}

	for _, test := range tests {
		vars, matched := template.Match(test.path, "")
		t.Logf("Path: %s, Matched: %v, Vars: %v (%s)", test.path, matched, vars, test.name)

		if matched != test.expected {
			t.Errorf("Path %s: expected matched=%v, got matched=%v", test.path, test.expected, matched)
		}
	}
}
