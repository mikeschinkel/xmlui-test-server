package test

import (
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

// TestAPIDataTypes tests basic data type validation for path parameters.
// Covers tests 01-08 from TEST_PLAN.md:
//   - Integer type ({id:int})
//   - String type ({email:string})
//   - UUID type ({uuid:uuid})
//   - Slug type ({slug:slug})
//   - Boolean type ({active:boolean})
//   - Real/Decimal type ({rating:real}, {value:decimal})
//   - Date type ({birth_date:date})
//   - Alphanumeric type ({sensor_id:alphanumeric})
func TestAPIDataTypes(t *testing.T) {
	server := setupComprehensiveTestServer(t)
	defer server.Cleanup()

	tests := []testRequest{
		// Test 01: Integer type validation
		{
			name:           "01_basic_int_parameter_valid",
			method:         "GET",
			path:           "/api/users/1",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name", "slug"},
		},
		{
			name:           "01_basic_int_parameter_invalid_string",
			method:         "GET",
			path:           "/api/users/abc",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'id' expected an integer type but got 'abc'",
				Instance:      "/api/users/abc",
				Parameter:     "id",
				ExpectedType:  "integer",
				ReceivedValue: "abc",
				Location:      apiutil.PathLocation,
				Suggestion:    "Use an integer for 'id' like 123, for example: /api/users/123",
			},
		},
		{
			name:           "01_basic_int_parameter_invalid_decimal",
			method:         "GET",
			path:           "/api/users/12.34",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'id' expected an integer type but got '12.34'",
				Instance:      "/api/users/12.34",
				Parameter:     "id",
				ExpectedType:  "integer",
				ReceivedValue: "12.34",
				Location:      apiutil.PathLocation,
				Suggestion:    "Use an integer for 'id' like 123, for example: /api/users/123",
			},
		},

		// Test 02: String type (always valid)
		{
			name:           "02_string_parameter_valid",
			method:         "GET",
			path:           "/api/users/by-name/Alice%20Carter",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name"},
		},
		{
			name:           "02_string_parameter_special_chars",
			method:         "GET",
			path:           "/api/users/by-name/Bob%20Nguyen",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name"},
		},

		// Test 03: UUID type validation
		{
			name:           "03_uuid_parameter_valid",
			method:         "GET",
			path:           "/api/users/by-uuid/550e8400-e29b-41d4-a716-446655440000",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name"},
		},
		{
			name:           "03_uuid_parameter_invalid_format",
			method:         "GET",
			path:           "/api/users/by-uuid/not-a-valid-uuid",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'uuid' expected type 'uuid' but received 'not-a-valid-uuid'",
				Instance:      "/api/users/by-uuid/not-a-valid-uuid",
				Parameter:     "uuid",
				ExpectedType:  "uuid",
				ReceivedValue: "not-a-valid-uuid",
				Location:      apiutil.PathLocation,
			},
		},
		{
			name:           "03_uuid_parameter_missing_hyphens",
			method:         "GET",
			path:           "/api/users/by-uuid/550e8400e29b41d4a716446655440000",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'uuid' expected type 'uuid' but received '550e8400e29b41d4a716446655440000'",
				Instance:      "/api/users/by-uuid/550e8400e29b41d4a716446655440000",
				Parameter:     "uuid",
				ExpectedType:  "uuid",
				ReceivedValue: "550e8400e29b41d4a716446655440000",
				Location:      apiutil.PathLocation,
			},
		},

		// Test 04: Slug type validation
		{
			name:           "04_slug_parameter_valid",
			method:         "GET",
			path:           "/api/users/by-slug/alice-carter",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name"},
		},
		{
			name:           "04_slug_parameter_invalid_uppercase",
			method:         "GET",
			path:           "/api/users/by-slug/Alice-Carter",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'slug' expected type 'slug' but received 'Alice-Carter'",
				Instance:      "/api/users/by-slug/Alice-Carter",
				Parameter:     "slug",
				ExpectedType:  "slug",
				ReceivedValue: "Alice-Carter",
				Location:      apiutil.PathLocation,
			},
		},
		{
			name:           "04_slug_parameter_invalid_special_chars",
			method:         "GET",
			path:           "/api/users/by-slug/alice_carter",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'slug' expected type 'slug' but received 'alice_carter'",
				Instance:      "/api/users/by-slug/alice_carter",
				Parameter:     "slug",
				ExpectedType:  "slug",
				ReceivedValue: "alice_carter",
				Location:      apiutil.PathLocation,
			},
		},

		// Test 05: Boolean type validation
		{
			name:           "05_boolean_parameter_true",
			method:         "GET",
			path:           "/api/users/by-status/true",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name", "active"},
		},
		{
			name:           "05_boolean_parameter_false",
			method:         "GET",
			path:           "/api/users/by-status/false",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name", "active"},
		},
		{
			name:           "05_boolean_parameter_invalid_1",
			method:         "GET",
			path:           "/api/users/by-status/1",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'active' expected type 'boolean' but received '1'",
				Instance:      "/api/users/by-status/1",
				Parameter:     "active",
				ExpectedType:  "boolean",
				ReceivedValue: "1",
				Location:      apiutil.PathLocation,
			},
		},
		{
			name:           "05_boolean_parameter_invalid_yes",
			method:         "GET",
			path:           "/api/users/by-status/yes",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'active' expected type 'boolean' but received 'yes'",
				Instance:      "/api/users/by-status/yes",
				Parameter:     "active",
				ExpectedType:  "boolean",
				ReceivedValue: "yes",
				Location:      apiutil.PathLocation,
			},
		},

		// Test 06: Real/Decimal type validation
		{
			name:           "06_real_parameter_valid_int",
			method:         "GET",
			path:           "/api/users/by-rating/4",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name", "rating"},
		},
		{
			name:           "06_real_parameter_valid_decimal",
			method:         "GET",
			path:           "/api/users/by-rating/4.5",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name", "rating"},
		},
		{
			name:           "06_real_parameter_invalid_string",
			method:         "GET",
			path:           "/api/users/by-rating/abc",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'rating' expected type 'real' but received 'abc'",
				Instance:      "/api/users/by-rating/abc",
				Parameter:     "rating",
				ExpectedType:  "real",
				ReceivedValue: "abc",
				Location:      apiutil.PathLocation,
			},
		},

		// Test 07: Date type validation
		{
			name:           "07_date_parameter_valid",
			method:         "GET",
			path:           "/api/users/by-birth-date/1990-01-15",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name", "birth_date"},
		},
		{
			name:           "07_date_parameter_invalid_format",
			method:         "GET",
			path:           "/api/users/by-birth-date/15-01-1990",
			expectedStatus: 422,
			// Date format constraint violation
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.ConstraintViolationErrorType,
				Title:         "Constraint Violation",
				Status:        422,
				Detail:        "Parameter 'birth_date' failed constraint 'format[yyyy-mm-dd]': value '15-01-1990' does not match required format",
				Instance:      "/api/users/by-birth-date/15-01-1990",
				Parameter:     "birth_date",
				Constraint:    "format[yyyy-mm-dd]",
				ReceivedValue: "15-01-1990",
				Location:      apiutil.PathLocation,
			},
		},

		// Test 08: Alphanumeric type validation
		{
			name:           "08_alphanumeric_parameter_valid",
			method:         "GET",
			path:           "/api/sensors/ABC123",
			expectedStatus: 200,
			expectedFields: []string{"sensor_id", "value"},
		},
		{
			name:           "08_alphanumeric_parameter_invalid_hyphen",
			method:         "GET",
			path:           "/api/sensors/ABC-123",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'sensor_id' expected type 'alphanumeric' but received 'ABC-123'",
				Instance:      "/api/sensors/ABC-123",
				Parameter:     "sensor_id",
				ExpectedType:  "alphanumeric",
				ReceivedValue: "ABC-123",
				Location:      apiutil.PathLocation,
			},
		},
		{
			name:           "08_alphanumeric_parameter_invalid_space",
			method:         "GET",
			path:           "/api/sensors/ABC%20123",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:          rfc9457.InvalidParameterErrorType,
				Title:         "Invalid Parameter Type",
				Status:        422,
				Detail:        "Parameter 'sensor_id' expected type 'alphanumeric' but received 'ABC 123'",
				Instance:      "/api/sensors/ABC%20123",
				Parameter:     "sensor_id",
				ExpectedType:  "alphanumeric",
				ReceivedValue: "ABC 123",
				Location:      apiutil.PathLocation,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestRequest(t, server.BaseURL, tt)
		})
	}
}
