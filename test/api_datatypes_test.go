package test

import (
	"testing"

	"github.com/mikeschinkel/go-rfc9457"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiresp"
)

// TestAPIDataTypes tests basic data type validation for path parameters.
// Covers tests 01-08 from TEST_SPECIFICATION.md:
//   - Integer type ({id:int})
//   - String type ({email:string})
//   - UUID type ({uuid:uuid})
//   - Slug type ({slug:slug})
//   - Boolean type ({active:boolean})
//   - Real/Decimal type ({rating:real}, {value:decimal})
//   - Date type ({birth_date:date})
//   - Alphanumeric type ({sensor_id:alphanumeric})
func TestAPIDataTypes(t *testing.T) {
	server := SetupTestServer(t)
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
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'id' expected an integer type but got 'abc'",
				Instance: "/api/users/abc",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "id",
						ExpectedType:  "integer",
						ReceivedValue: "abc",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use an integer for 'id' like 123, for example: /api/users/123",
					},
				},
			},
		},
		{
			name:           "01_basic_int_parameter_invalid_decimal",
			method:         "GET",
			path:           "/api/users/12.34",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'id' expected an integer type but got '12.34'",
				Instance: "/api/users/12.34",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "id",
						ExpectedType:  "integer",
						ReceivedValue: "12.34",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use an integer for 'id' like 123, for example: /api/users/123",
					},
				},
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
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'uuid' expected a uuid type but got 'not-a-valid-uuid'",
				Instance: "/api/users/by-uuid/not-a-valid-uuid",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "uuid",
						ExpectedType:  "uuid",
						ReceivedValue: "not-a-valid-uuid",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use a uuid for 'uuid' like f81d4fae-7dec-11d0-a765-00a0c91e6bf6, for example: /api/users/by-uuid/deadbeef-cafe-4011-8123-b1d5c0d51234",
					},
				},
			},
		},
		{
			name:           "03_uuid_parameter_missing_hyphens",
			method:         "GET",
			path:           "/api/users/by-uuid/550e8400e29b41d4a716446655440000",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'uuid' expected a uuid type but got '550e8400e29b41d4a716446655440000'",
				Instance: "/api/users/by-uuid/550e8400e29b41d4a716446655440000",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "uuid",
						ExpectedType:  "uuid",
						ReceivedValue: "550e8400e29b41d4a716446655440000",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use a uuid for 'uuid' like f81d4fae-7dec-11d0-a765-00a0c91e6bf6, for example: /api/users/by-uuid/deadbeef-cafe-4011-8123-b1d5c0d51234",
					},
				},
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
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'slug' expected a slug type but got 'Alice-Carter'",
				Instance: "/api/users/by-slug/Alice-Carter",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "slug",
						ExpectedType:  "slug",
						ReceivedValue: "Alice-Carter",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use a slug for 'slug' like abc-123, for example: /api/users/by-slug/abc-123",
					},
				},
			},
		},
		{
			name:           "04_slug_parameter_invalid_special_chars",
			method:         "GET",
			path:           "/api/users/by-slug/alice_carter",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'slug' expected a slug type but got 'alice_carter'",
				Instance: "/api/users/by-slug/alice_carter",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "slug",
						ExpectedType:  "slug",
						ReceivedValue: "alice_carter",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use a slug for 'slug' like abc-123, for example: /api/users/by-slug/abc-123",
					},
				},
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
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'active' expected a boolean type but got '1'",
				Instance: "/api/users/by-status/1",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "active",
						ExpectedType:  "boolean",
						ReceivedValue: "1",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use a boolean for 'active' like true, for example: /api/users/by-status/true",
					},
				},
			},
		},
		{
			name:           "05_boolean_parameter_invalid_yes",
			method:         "GET",
			path:           "/api/users/by-status/yes",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'active' expected a boolean type but got 'yes'",
				Instance: "/api/users/by-status/yes",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "active",
						ExpectedType:  "boolean",
						ReceivedValue: "yes",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use a boolean for 'active' like true, for example: /api/users/by-status/true",
					},
				},
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
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'rating' expected a real type but got 'abc'",
				Instance: "/api/users/by-rating/abc",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "rating",
						ExpectedType:  "real",
						ReceivedValue: "abc",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use a real for 'rating' like 1.2345, for example: /api/users/by-rating/1.2345",
					},
				},
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
			// Date type validation error (value fails both constraint AND type validation)
			// Per validation logic: when format constraint fails, we check if it's a valid date
			// Since '15-01-1990' is not yyyy-mm-dd format, it fails date type validation
			// Therefore we return type error instead of constraint error
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'birth_date' expected a date type but got '15-01-1990'",
				Instance: "/api/users/by-birth-date/15-01-1990",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "birth_date",
						ExpectedType:  "date",
						ReceivedValue: "15-01-1990",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use a date for 'birth_date' like 1999-12-31, for example: /api/users/by-birth-date/1999-12-31",
					},
				},
			},
		},

		// Test 08: Alphanumeric type validation
		{
			name:           "08_alphanumeric_parameter_valid",
			method:         "GET",
			path:           "/api/measurements/by-sensor/ABC123",
			expectedStatus: 200,
			expectedFields: []string{"sensor_id", "value"},
		},
		{
			name:           "08_alphanumeric_parameter_invalid_hyphen",
			method:         "GET",
			path:           "/api/measurements/by-sensor/ABC-123",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'sensor_id' expected an alphanumeric type but got 'ABC-123'",
				Instance: "/api/measurements/by-sensor/ABC-123",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "sensor_id",
						ExpectedType:  "alphanumeric",
						ReceivedValue: "ABC-123",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use an alphanumeric for 'sensor_id' like abc123, for example: /api/measurements/by-sensor/abc123",
					},
				},
			},
		},
		{
			name:           "08_alphanumeric_parameter_invalid_space",
			method:         "GET",
			path:           "/api/measurements/by-sensor/ABC%20123",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'sensor_id' expected an alphanumeric type but got 'ABC 123'",
				Instance: "/api/measurements/by-sensor/ABC%20123",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "sensor_id",
						ExpectedType:  "alphanumeric",
						ReceivedValue: "ABC 123",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use an alphanumeric for 'sensor_id' like abc123, for example: /api/measurements/by-sensor/abc123",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestRequest(t, server.BaseURL, tt)
		})
	}
}
