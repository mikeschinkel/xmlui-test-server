package test

import (
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiresp"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

// TestAPIQueryParams tests query parameter extraction and validation.
// Covers tests 20-22 from OLD_TEST_FILE_ANALYSIS.md:
//   - Test 20: Basic query parameter extraction
//   - Test 21: Optional query parameters with default values
//   - Test 22: Combined path + query parameters
//
// This tests the critical functionality of:
//   - Extracting query parameters from URL query string
//   - Applying default values to optional parameters
//   - Combining path and query parameters in SQL execution
//   - Validating query parameter types and constraints
//   - RFC 9457 error responses for invalid query parameters
func TestAPIQueryParams(t *testing.T) {
	server := setupComprehensiveTestServer(t)
	defer server.Cleanup()

	tests := []testRequest{
		// Test 20: Basic query parameter extraction
		{
			name:           "20_query_only_endpoint_all_params_provided",
			method:         "GET",
			path:           "/api/users/search?email=alice&active=true&min_score=50&limit=5",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name", "score", "active"},
		},
		{
			name:           "20_query_param_email_required",
			method:         "GET",
			path:           "/api/users/search?email=bob@example.com&active=true&min_score=0&limit=10",
			expectedStatus: 200,
			expectedFields: []string{"id", "email", "name", "score", "active"},
		},
		{
			name:           "20_query_param_invalid_integer_type",
			method:         "GET",
			path:           "/api/users/search?email=alice&active=true&min_score=abc&limit=5",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'min_score' expected an integer type but got 'abc'",
				Instance: "/api/users/search?email=alice&active=true&min_score=abc&limit=5",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "min_score",
						ExpectedType:  "integer",
						ReceivedValue: "abc",
						Location:      apiresp.QueryLocation,
						Suggestion:    "Use an integer for 'min_score' like 123, for example: /api/users/search?email={EMAIL}&active={ACTIVE}&limit={LIMIT}&min_score=123",
					},
				},
			},
		},
		{
			name:           "20_query_param_invalid_boolean_type",
			method:         "GET",
			path:           "/api/users/search?email=alice&active=yes&min_score=0&limit=5",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'active' expected a boolean type but got 'yes'",
				Instance: "/api/users/search?email=alice&active=yes&min_score=0&limit=5",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "active",
						ExpectedType:  "boolean",
						ReceivedValue: "yes",
						Location:      apiresp.QueryLocation,
						Suggestion:    "Use a boolean for 'active' like true, for example: /api/users/search?email={EMAIL}&min_score={MIN_SCORE}&limit={LIMIT}&active=true",
					},
				},
			},
		},
		{
			name:           "20_query_param_constraint_violation_range",
			method:         "GET",
			path:           "/api/users/search?email=alice&active=true&min_score=0&limit=500",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.ConstraintViolationErrorType,
				Title:    "Constraint Violation",
				Status:   422,
				Detail:   "Parameter 'limit' with value '500' failed constraint validation: value 500 is outside the allowed range of 1..100",
				Instance: "/api/users/search?email=alice&active=true&min_score=0&limit=500",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "limit",
						ExpectedType:  "integer",
						ReceivedValue: "500",
						Location:      apiresp.QueryLocation,
						Constraint:    "range[1..100]",
						Suggestion:    "Ensure parameter 'limit' satisfies the constraint: range[1..100], for example: /api/users/search?email={EMAIL}&active={ACTIVE}&min_score={MIN_SCORE}&limit=50",
					},
				},
			},
		},
		{
			name:           "20_query_param_notempty_constraint_violation",
			method:         "GET",
			path:           "/api/users/search?email=&active=true&min_score=0&limit=10",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.ConstraintViolationErrorType,
				Title:    "Constraint Violation",
				Status:   422,
				Detail:   "Parameter 'email' with value '' failed constraint validation: value cannot be empty",
				Instance: "/api/users/search?email=&active=true&min_score=0&limit=10",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "email",
						ExpectedType:  "string",
						ReceivedValue: "",
						Location:      apiresp.QueryLocation,
						Constraint:    "notempty",
						Suggestion:    "Ensure parameter 'email' satisfies the constraint: notempty, for example: /api/users/search?active={ACTIVE}&min_score={MIN_SCORE}&limit={LIMIT}&email=example",
					},
				},
			},
		},

		// Test 21: Optional query parameters with default values
		{
			name:           "21_optional_query_params_all_defaults",
			method:         "GET",
			path:           "/api/projects/search",
			expectedStatus: 200,
			expectedFields: []string{"id", "name", "status", "budget"},
		},
		{
			name:           "21_optional_query_params_some_defaults",
			method:         "GET",
			path:           "/api/projects/search?name=Test",
			expectedStatus: 200,
			expectedFields: []string{"id", "name", "status", "budget"},
		},
		{
			name:           "21_optional_query_params_no_defaults",
			method:         "GET",
			path:           "/api/projects/search?name=Project&status=archived&min_budget=5000",
			expectedStatus: 200,
			expectedFields: []string{"id", "name", "status", "budget"},
		},
		{
			name:           "21_optional_query_param_invalid_enum",
			method:         "GET",
			path:           "/api/projects/search?status=invalid",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.ConstraintViolationErrorType,
				Title:    "Constraint Violation",
				Status:   422,
				Detail:   "Parameter 'status' with value 'invalid' failed constraint validation: value 'invalid' is not in the allowed set: [active, archived, draft]",
				Instance: "/api/projects/search?status=invalid",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "status",
						ExpectedType:  "string",
						ReceivedValue: "invalid",
						Location:      apiresp.QueryLocation,
						Constraint:    "enum[active,archived,draft]",
						Suggestion:    "Ensure parameter 'status' satisfies the constraint: enum[active,archived,draft], for example: /api/projects/search?status=active",
					},
				},
			},
		},
		{
			name:           "21_optional_query_param_invalid_decimal",
			method:         "GET",
			path:           "/api/projects/search?min_budget=invalid",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'min_budget' expected a decimal type but got 'invalid'",
				Instance: "/api/projects/search?min_budget=invalid",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "min_budget",
						ExpectedType:  "decimal",
						ReceivedValue: "invalid",
						Location:      apiresp.QueryLocation,
						Suggestion:    "Use a decimal for 'min_budget' like 1.2345, for example: /api/projects/search?min_budget=1.2345",
					},
				},
			},
		},

		// Test 22: Combined path + query parameters
		{
			name:           "22_path_and_query_params_valid",
			method:         "GET",
			path:           "/api/tasks/search/1?q=urgent&limit=10&offset=0&sort=desc",
			expectedStatus: 200,
			expectedFields: []string{"id", "title", "status", "priority", "assignee_email"},
		},
		{
			name:           "22_path_param_invalid_with_query_params",
			method:         "GET",
			path:           "/api/tasks/search/abc?q=test&limit=10&offset=0&sort=asc",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.InvalidURLParameterErrorType,
				Title:    "Invalid URL Parameter",
				Status:   422,
				Detail:   "Parameter 'project_id' expected an integer type but got 'abc'",
				Instance: "/api/tasks/search/abc?q=test&limit=10&offset=0&sort=asc",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "project_id",
						ExpectedType:  "integer",
						ReceivedValue: "abc",
						Location:      apiresp.PathLocation,
						Suggestion:    "Use an integer for 'project_id' like 123, for example: /api/tasks/search/123?q={Q}&limit={LIMIT}&offset={OFFSET}&sort={SORT}",
					},
				},
			},
		},
		{
			name:           "22_query_param_invalid_with_path_param",
			method:         "GET",
			path:           "/api/tasks/search/1?q=test&limit=500&offset=0&sort=asc",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.ConstraintViolationErrorType,
				Title:    "Constraint Violation",
				Status:   422,
				Detail:   "Parameter 'limit' with value '500' failed constraint validation: value 500 is outside the allowed range of 1..100",
				Instance: "/api/tasks/search/1?q=test&limit=500&offset=0&sort=asc",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "limit",
						ExpectedType:  "integer",
						ReceivedValue: "500",
						Location:      apiresp.QueryLocation,
						Constraint:    "range[1..100]",
						Suggestion:    "Ensure parameter 'limit' satisfies the constraint: range[1..100], for example: /api/tasks/search/{PROJECT_ID}?q={Q}&offset={OFFSET}&sort={SORT}&limit=50",
					},
				},
			},
		},
		{
			name:           "22_query_param_enum_constraint_violation",
			method:         "GET",
			path:           "/api/tasks/search/1?q=test&limit=10&offset=0&sort=invalid",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.ConstraintViolationErrorType,
				Title:    "Constraint Violation",
				Status:   422,
				Detail:   "Parameter 'sort' with value 'invalid' failed constraint validation: value 'invalid' is not in the allowed set: [asc, desc]",
				Instance: "/api/tasks/search/1?q=test&limit=10&offset=0&sort=invalid",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "sort",
						ExpectedType:  "string",
						ReceivedValue: "invalid",
						Location:      apiresp.QueryLocation,
						Constraint:    "enum[asc,desc]",
						Suggestion:    "Ensure parameter 'sort' satisfies the constraint: enum[asc,desc], for example: /api/tasks/search/{PROJECT_ID}?q={Q}&limit={LIMIT}&offset={OFFSET}&sort=asc",
					},
				},
			},
		},
		{
			name:           "22_query_param_offset_range_violation",
			method:         "GET",
			path:           "/api/tasks/search/1?q=test&limit=10&offset=2000&sort=asc",
			expectedStatus: 422,
			expectedRFC9457: &rfc9457.Response{
				Type:     rfc9457.ConstraintViolationErrorType,
				Title:    "Constraint Violation",
				Status:   422,
				Detail:   "Parameter 'offset' with value '2000' failed constraint validation: value 2000 is outside the allowed range of 0..1000",
				Instance: "/api/tasks/search/1?q=test&limit=10&offset=2000&sort=asc",
				Extensions: []rfc9457.Extension{
					apiresp.RFC9457Extension{
						Parameter:     "offset",
						ExpectedType:  "integer",
						ReceivedValue: "2000",
						Location:      apiresp.QueryLocation,
						Constraint:    "range[0..1000]",
						Suggestion:    "Ensure parameter 'offset' satisfies the constraint: range[0..1000], for example: /api/tasks/search/{PROJECT_ID}?q={Q}&limit={LIMIT}&sort={SORT}&offset=500",
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
