package apiutil_test

import (
	"encoding/json"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiutil"
)

// assertExtensionEqual compares two RFC9457Extension structs field by field
func assertExtensionEqual(t *testing.T, got, want *apiutil.RFC9457Extension) {
	t.Helper()

	if got.Parameter != want.Parameter {
		t.Errorf("Parameter: got %v, want %v", got.Parameter, want.Parameter)
	}
	if got.ExpectedType != want.ExpectedType {
		t.Errorf("ExpectedType: got %v, want %v", got.ExpectedType, want.ExpectedType)
	}
	if got.ReceivedValue != want.ReceivedValue {
		t.Errorf("ReceivedValue: got %v, want %v", got.ReceivedValue, want.ReceivedValue)
	}
	if got.Location != want.Location {
		t.Errorf("Location: got %v, want %v", got.Location, want.Location)
	}
	if got.Suggestion != want.Suggestion {
		t.Errorf("Suggestion: got %v, want %v", got.Suggestion, want.Suggestion)
	}
	if got.Constraint != want.Constraint {
		t.Errorf("Constraint: got %v, want %v", got.Constraint, want.Constraint)
	}

	// Compare ValidationErrors slice
	if len(got.ValidationErrors) != len(want.ValidationErrors) {
		t.Errorf("ValidationErrors length: got %d, want %d", len(got.ValidationErrors), len(want.ValidationErrors))
	} else {
		for i := range want.ValidationErrors {
			gotErr := got.ValidationErrors[i]
			wantErr := want.ValidationErrors[i]
			if gotErr.Parameter != wantErr.Parameter {
				t.Errorf("ValidationErrors[%d].Parameter: got %q, want %q", i, gotErr.Parameter, wantErr.Parameter)
			}
			if gotErr.Location != wantErr.Location {
				t.Errorf("ValidationErrors[%d].Location: got %v, want %v", i, gotErr.Location, wantErr.Location)
			}
			if gotErr.Expected != wantErr.Expected {
				t.Errorf("ValidationErrors[%d].Expected: got %q, want %q", i, gotErr.Expected, wantErr.Expected)
			}
			if gotErr.Received != wantErr.Received {
				t.Errorf("ValidationErrors[%d].Received: got %q, want %q", i, gotErr.Received, wantErr.Received)
			}
			if gotErr.Message != wantErr.Message {
				t.Errorf("ValidationErrors[%d].Message: got %q, want %q", i, gotErr.Message, wantErr.Message)
			}
		}
	}
}

func TestRFC9457Extension_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		ext  *apiutil.RFC9457Extension
		want string
	}{
		{
			name: "extension_all_fields",
			ext: &apiutil.RFC9457Extension{
				Parameter:     "id",
				ExpectedType:  "int",
				ReceivedValue: "abc",
				Location:      apiutil.PathLocation,
			},
			want: `{"parameter":"id","expected_type":"int","received_value":"abc","location":"path"}`,
		},
		{
			name: "extension_with_constraint",
			ext: &apiutil.RFC9457Extension{
				Parameter:     "score",
				ReceivedValue: "150",
				Location:      apiutil.PathLocation,
				Constraint:    "range[0..100]",
			},
			want: `{"parameter":"score","received_value":"150","location":"path","constraint":"range[0..100]"}`,
		},
		{
			name: "extension_with_suggestion",
			ext: &apiutil.RFC9457Extension{
				Parameter:  "date",
				Location:   apiutil.PathLocation,
				Suggestion: "Try using format: 1990-05-15",
			},
			want: `{"parameter":"date","location":"path","suggestion":"Try using format: 1990-05-15"}`,
		},
		{
			name: "extension_with_validation_errors",
			ext: &apiutil.RFC9457Extension{
				ValidationErrors: []apiutil.ValidationError{
					{
						Parameter: "email",
						Location:  apiutil.BodyLocation,
						Expected:  "valid email",
						Received:  "invalid@",
						Message:   "Email format is invalid",
					},
				},
			},
			want: `{"validation_errors":[{"parameter":"email","location":"body","expected":"valid email","received":"invalid@","message":"Email format is invalid"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.ext)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("Marshal mismatch:\ngot:  %s\nwant: %s", string(got), tt.want)
			}
		})
	}
}

func TestRFC9457Extension_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    *apiutil.RFC9457Extension
		wantErr bool
	}{
		{
			name: "extension_roundtrip",
			json: `{"parameter":"id","expected_type":"int","received_value":"abc","location":"path"}`,
			want: &apiutil.RFC9457Extension{
				Parameter:     "id",
				ExpectedType:  "int",
				ReceivedValue: "abc",
				Location:      apiutil.PathLocation,
			},
		},
		{
			name: "extension_with_validation_errors",
			json: `{"validation_errors":[{"parameter":"email","location":"body","expected":"valid email","received":"invalid@","message":"Email format is invalid"}]}`,
			want: &apiutil.RFC9457Extension{
				ValidationErrors: []apiutil.ValidationError{
					{
						Parameter: "email",
						Location:  apiutil.BodyLocation,
						Expected:  "valid email",
						Received:  "invalid@",
						Message:   "Email format is invalid",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got apiutil.RFC9457Extension
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error: %v, wantErr: %v", err, tt.wantErr)
			}

			assertExtensionEqual(t, &got, tt.want)
		})
	}
}

func TestRFC9457Extension_SetSuggestion(t *testing.T) {
	ext := &apiutil.RFC9457Extension{
		Parameter: "date",
		Location:  apiutil.PathLocation,
	}

	ext.SetSuggestion("Try using format: YYYY-MM-DD")

	if ext.Suggestion != "Try using format: YYYY-MM-DD" {
		t.Errorf("SetSuggestion: got %q, want %q", ext.Suggestion, "Try using format: YYYY-MM-DD")
	}
}

func TestRFC9457Extension_SetConstraint(t *testing.T) {
	ext := &apiutil.RFC9457Extension{
		Parameter: "score",
		Location:  apiutil.PathLocation,
	}

	ext.SetConstraint("range[0..100]")

	if ext.Constraint != "range[0..100]" {
		t.Errorf("SetConstraint: got %v, want %q", ext.Constraint, "range[0..100]")
	}
}

func TestRFC9457Extension_AddValidationError(t *testing.T) {
	ext := &apiutil.RFC9457Extension{
		Parameter: "user",
		Location:  apiutil.BodyLocation,
	}

	// Add first validation error
	ve1 := apiutil.ValidationError{
		Parameter: "email",
		Location:  apiutil.BodyLocation,
		Expected:  "valid email",
		Received:  "invalid@",
		Message:   "Email format is invalid",
	}
	ext.AddValidationError(ve1)

	if len(ext.ValidationErrors) != 1 {
		t.Fatalf("ValidationErrors length: got %d, want 1", len(ext.ValidationErrors))
	}

	got := ext.ValidationErrors[0]
	if got.Parameter != "email" {
		t.Errorf("Parameter: got %q, want %q", got.Parameter, "email")
	}
	if got.Location != apiutil.BodyLocation {
		t.Errorf("Location: got %v, want %v", got.Location, apiutil.BodyLocation)
	}
	if got.Expected != "valid email" {
		t.Errorf("Expected: got %q, want %q", got.Expected, "valid email")
	}
	if got.Received != "invalid@" {
		t.Errorf("Received: got %q, want %q", got.Received, "invalid@")
	}
	if got.Message != "Email format is invalid" {
		t.Errorf("Message: got %q, want %q", got.Message, "Email format is invalid")
	}

	// Add second validation error
	ve2 := apiutil.ValidationError{
		Parameter: "age",
		Location:  apiutil.BodyLocation,
		Expected:  "integer",
		Received:  "abc",
		Message:   "Age must be numeric",
	}
	ext.AddValidationError(ve2)

	if len(ext.ValidationErrors) != 2 {
		t.Fatalf("ValidationErrors length after second add: got %d, want 2", len(ext.ValidationErrors))
	}
}

func TestRFC9457Extension_MultipleMutations(t *testing.T) {
	// Test that multiple mutations work correctly
	ext := apiutil.NewRFC9457Extension(apiutil.RFC9457ExtensionArgs{
		Parameter: "score",
		Location:  apiutil.PathLocation,
	})

	ext.SetConstraint("range[0..100]")
	ext.SetSuggestion("Provide a value between 0 and 100")
	ext.AddValidationError(apiutil.ValidationError{
		Parameter: "score",
		Location:  apiutil.PathLocation,
		Expected:  "0-100",
		Received:  "150",
		Message:   "Score out of range",
	})
	ext.AddValidationError(apiutil.ValidationError{
		Parameter: "rating",
		Location:  apiutil.PathLocation,
		Expected:  "0.0-5.0",
		Received:  "6.5",
		Message:   "Rating out of range",
	})

	// Verify all fields were set
	if ext.Constraint != "range[0..100]" {
		t.Errorf("Constraint not set")
	}
	if ext.Suggestion != "Provide a value between 0 and 100" {
		t.Errorf("Suggestion not set")
	}
	if len(ext.ValidationErrors) != 2 {
		t.Errorf("ValidationErrors not set: got %d, want 2", len(ext.ValidationErrors))
	}
}

func TestNewRFC9457Extension(t *testing.T) {
	args := apiutil.RFC9457ExtensionArgs{
		Parameter:     "test_param",
		ExpectedType:  "int",
		ReceivedValue: "abc",
		Location:      apiutil.PathLocation,
		Constraint:    "range[0..10]",
		Suggestion:    "Use a number",
		ValidationErrors: []apiutil.ValidationError{
			{Parameter: "test", Location: apiutil.PathLocation, Expected: "int", Received: "abc", Message: "Invalid"},
		},
	}

	got := apiutil.NewRFC9457Extension(args)

	// Build expected result from args
	want := &apiutil.RFC9457Extension{
		Parameter:        args.Parameter,
		ExpectedType:     args.ExpectedType,
		ReceivedValue:    args.ReceivedValue,
		Location:         args.Location,
		Constraint:       args.Constraint,
		Suggestion:       args.Suggestion,
		ValidationErrors: args.ValidationErrors,
	}

	assertExtensionEqual(t, got, want)
}

func TestRFC9457Extension_Error(t *testing.T) {
	ext := &apiutil.RFC9457Extension{
		Parameter:     "id",
		ExpectedType:  "int",
		ReceivedValue: "abc",
		Location:      apiutil.PathLocation,
		Constraint:    "range[0..100]",
		Suggestion:    "Use a valid integer",
	}

	errStr := ext.Error()

	// Verify the error string contains key information
	if errStr == "" {
		t.Error("Error() returned empty string")
	}

	// The error string should be formatted as a multi-line description
	// We just verify it's not empty for now, as the exact format may change
}
