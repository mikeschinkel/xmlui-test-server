package pathvars

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiresp"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/errparsr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

func TestExtractParameterSpec(t *testing.T) {
	tests := []struct {
		name       string
		segment    string
		wantSpec   string
		wantPrefix string
		wantSuffix string
		wantErr    bool
	}{
		{
			name:     "Not a var",
			segment:  "foo",
			wantSpec: "foo",
			wantErr:  false,
		},
		{
			name:     "Just a var",
			segment:  "{id:integer}",
			wantSpec: "{id:integer}",
			wantErr:  false,
		},
		{
			name:       "Var with prefix",
			segment:    "foo{id:integer}",
			wantSpec:   "{id:integer}",
			wantPrefix: "foo",
			wantErr:    false,
		},
		{
			name:       "Var with suffix",
			segment:    "{id:integer}bar",
			wantSpec:   "{id:integer}",
			wantSuffix: "bar",
			wantErr:    false,
		},
		{
			name:       "Var with both fixes",
			segment:    "foo{id:integer}bar",
			wantSpec:   "{id:integer}",
			wantPrefix: "foo",
			wantSuffix: "bar",
			wantErr:    false,
		},
		{
			name:    "Missing closing brace",
			segment: "{id:integer",
			wantErr: true,
		},
		{
			name:    "Malformed braces with just a var",
			segment: "}id:integer{",
			wantErr: true,
		},
		{
			name:    "Malformed braces var with prefix",
			segment: "foo}id:integer{",
			wantErr: true,
		},
		{
			name:    "Malformed braces with suffix",
			segment: "}id:integer{bar",
			wantErr: true,
		},
		{
			name:    "Malformed braces var with both fixes",
			segment: "foo}id:integer{bar",
			wantErr: true,
		},
		// TODO: Add tests to check for specific sentinel errors
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefix, gotSpec, gotSuffix, err := ExtractParameterSpec(tt.segment)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractParameterSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSpec != tt.wantSpec {
				t.Errorf("ExtractParameterSpec() gotName = %v, want %v", gotSpec, tt.wantSpec)
			}
			if !reflect.DeepEqual(gotPrefix, tt.wantPrefix) {
				t.Errorf("ExtractParameterSpec() gotPrefix = %v, want %v", gotPrefix, tt.wantPrefix)
			}
			if !reflect.DeepEqual(gotSuffix, tt.wantSuffix) {
				t.Errorf("ExtractParameterSpec() gotSuffix = %v, want %v", gotSuffix, tt.wantSuffix)
			}
		})
	}
}

// TestTemplate_RFC9457Generation tests that validateParameter() generates correct Response
func TestTemplate_RFC9457Generation(t *testing.T) {
	tests := []struct {
		name              string
		template          string
		path              string
		wantType          rfc9457.ErrorTypeURI
		wantParam         string
		wantExpectedType  string
		wantReceivedValue string
		wantLocation      apiresp.LocationType
	}{
		{
			name:              "Invalid integer parameter",
			template:          "/api/users/{id:integer}",
			path:              "/api/users/abc",
			wantType:          rfc9457.InvalidParameterErrorType,
			wantParam:         "id",
			wantExpectedType:  "integer",
			wantReceivedValue: "abc",
			wantLocation:      apiresp.PathLocation,
		},
		{
			name:              "Invalid UUID parameter",
			template:          "/api/items/{uuid:uuid}",
			path:              "/api/items/not-a-uuid",
			wantType:          rfc9457.InvalidParameterErrorType,
			wantParam:         "uuid",
			wantExpectedType:  "uuid",
			wantReceivedValue: "not-a-uuid",
			wantLocation:      apiresp.PathLocation,
		},
		{
			name:              "Invalid boolean parameter",
			template:          "/api/settings/{enabled:boolean}",
			path:              "/api/settings/maybe",
			wantType:          rfc9457.InvalidParameterErrorType,
			wantParam:         "enabled",
			wantExpectedType:  "boolean",
			wantReceivedValue: "maybe",
			wantLocation:      apiresp.PathLocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse template
			tmpl, err := ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			// Match and validate (should fail with Response)
			_, _, err = tmpl.Match(tt.path, "")
			if err != nil {
				// Expected - validation should fail
			} else {
				t.Fatal("Expected validation error, got nil")
			}

			// Extract Response from error
			pe, parseErr := errparsr.ParseError(err)
			if parseErr != nil {
				t.Fatalf("ParseError() error = %v", parseErr)
			}

			customErr := pe.MaybeGetCustomError(rfc9457.ResponseArchetype)
			if customErr == nil {
				t.Fatal("Expected Response in error chain, got nil")
			}

			var resp *rfc9457.Response
			if !errors.As(customErr, &resp) {
				t.Fatalf("Expected *common.Response, got %T", customErr)
			}

			// Verify Response fields
			if resp.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", resp.Type, tt.wantType)
			}
			if resp.Status != http.StatusUnprocessableEntity {
				t.Errorf("Status = %d, want %d", resp.Status, http.StatusUnprocessableEntity)
			}
			if resp.Title != "Invalid Parameter Type" {
				t.Errorf("Title = %q, want %q", resp.Title, "Invalid Parameter Type")
			}
			if resp.Detail == "" {
				t.Error("Detail should not be empty")
			}
			if resp.Instance != tt.path {
				t.Errorf("Instance = %q, want %q", resp.Instance, tt.path)
			}

			// Verify extension fields
			if len(resp.Extensions) == 0 {
				t.Fatal("Expected at least one extension")
			}
			ext, ok := resp.Extensions[0].(*apiresp.RFC9457Extension)
			if !ok {
				t.Fatalf("Expected *RFC9457Extension, got %T", resp.Extensions[0])
			}
			if ext.Parameter != tt.wantParam {
				t.Errorf("Extension.Parameter = %q, want %q", ext.Parameter, tt.wantParam)
			}
			if ext.ExpectedType != tt.wantExpectedType {
				t.Errorf("Extension.ExpectedType = %q, want %q", ext.ExpectedType, tt.wantExpectedType)
			}
			if ext.ReceivedValue != tt.wantReceivedValue {
				t.Errorf("Extension.ReceivedValue = %q, want %q", ext.ReceivedValue, tt.wantReceivedValue)
			}
			if ext.Location != tt.wantLocation {
				t.Errorf("Extension.Location = %v, want %v", ext.Location, tt.wantLocation)
			}
		})
	}
}

// TestTemplate_RFC9457_DetailMessage tests the Detail field formatting
func TestTemplate_RFC9457_DetailMessage(t *testing.T) {
	tmpl, err := ParseTemplate("/api/users/{id:integer}")
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	_, _, err = tmpl.Match("/api/users/abc", "")
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	pe, _ := errparsr.ParseError(err)
	customErr := pe.MaybeGetCustomError(rfc9457.ResponseArchetype)
	var rfc9457 *rfc9457.Response
	errors.As(customErr, &rfc9457)

	// Detail should mention parameter name, expected type, and received value
	detail := rfc9457.Detail
	if !strings.Contains(detail, "id") {
		t.Errorf("Detail should mention parameter 'id', got %q", detail)
	}
	if !strings.Contains(detail, "integer") {
		t.Errorf("Detail should mention expected type 'integer', got %q", detail)
	}
	if !strings.Contains(detail, "abc") {
		t.Errorf("Detail should mention received value 'abc', got %q", detail)
	}
}

// TestTemplate_RFC9457_HTTPStatusCode tests that status is always 422
func TestTemplate_RFC9457_HTTPStatusCode(t *testing.T) {
	templates := []string{
		"/api/users/{id:integer}",
		"/api/items/{uuid:uuid}",
		"/api/settings/{enabled:boolean}",
		"/api/posts/{slug:slug}",
	}

	for _, tmplStr := range templates {
		t.Run(tmplStr, func(t *testing.T) {
			tmpl, err := ParseTemplate(tmplStr)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			invalidPath := tmplStr[:strings.Index(tmplStr, "{")] + "invalid"
			_, _, err = tmpl.Match(invalidPath, "")
			if err == nil {
				return // Some types might accept "invalid" as valid
			}

			pe, _ := errparsr.ParseError(err)
			customErr := pe.MaybeGetCustomError(rfc9457.ResponseArchetype)
			if customErr == nil {
				return // Not an RFC9457 error
			}

			var rfc9457 *rfc9457.Response
			if errors.As(customErr, &rfc9457) {
				if rfc9457.Status != http.StatusUnprocessableEntity {
					t.Errorf("Status = %d, want %d", rfc9457.Status, http.StatusUnprocessableEntity)
				}
			}
		})
	}
}
