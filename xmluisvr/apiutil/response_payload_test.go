package apiutil_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

// TestInternalServerErrorPayload tests the InternalServerErrorPayload function
func TestInternalServerErrorPayload(t *testing.T) {
	var resp *rfc9457.Response

	req := httptest.NewRequest("GET", "/api/users/abc", nil)

	payload := apiutil.InternalServerErrorPayload(req, apiutil.PayloadArgs{})

	// Should return Response
	if !errors.As(payload, &resp) {
		t.Fatalf("Expected *apiutil.Response, got %T", payload)
	}

	// Verify all fields
	if resp.Type != rfc9457.InternalServerErrorType {
		t.Errorf("Type: got %v, want %v", resp.Type, rfc9457.InternalServerErrorType)
	}
	if resp.Title != "Internal Server Error" {
		t.Errorf("Title: got %q, want %q", resp.Title, "Internal Server Error")
	}
	if resp.Status != http.StatusInternalServerError {
		t.Errorf("Status: got %d, want %d", resp.Status, http.StatusInternalServerError)
	}
	if resp.Detail == "" {
		t.Error("Detail should not be empty")
	}
	if resp.Instance != req.RequestURI {
		t.Errorf("Instance: got %q, want %q", resp.Instance, req.RequestURI)
	}

	if len(resp.Extensions) != 1 {
		t.Errorf("Number of extensions should be 1, got %d", len(resp.Extensions))
	}
	respExt := resp.Extensions[0]

	ext, ok := respExt.(apiutil.RFC9457Extension)
	if !ok {
		t.Errorf("Extensions is not of type apiutil.RFC9457Extension, got %T instead", respExt)
	}

	// Optional fields should be empty
	if ext.Parameter != "" {
		t.Errorf("Parameter should be empty, got %q", ext.Parameter)
	}
	if ext.ExpectedType != "" {
		t.Errorf("ExpectedType should be empty, got %q", ext.ExpectedType)
	}
	if ext.ReceivedValue != "" {
		t.Errorf("ReceivedValue should be empty, got %q", ext.ReceivedValue)
	}
	if ext.Location != "" {
		t.Errorf("Location should be empty, got %q", ext.Location)
	}
	if ext.Constraint != nil {
		t.Errorf("Constraint should be nil, got %v", ext.Constraint)
	}
	if ext.Suggestion != "" {
		t.Errorf("Suggestion should be empty, got %q", ext.Suggestion)
	}
	if len(ext.ValidationErrors) != 0 {
		t.Errorf("ValidationErrors should be empty, got %d items", len(ext.ValidationErrors))
	}
}

// TestEndpointNotMatchedPayload tests the EndpointNotMatchedPayload function
func TestEndpointNotMatchedPayload(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/nonexistent", nil)

	payload := apiutil.EndpointNotMatchedPayload(req, apiutil.PayloadArgs{})
	// Should return Response

	var resp *rfc9457.Response
	if !errors.As(payload, &resp) {
		t.Fatalf("Expected *apiutil.Response, got %T", payload)
	}

	// Verify all fields
	if resp.Type != rfc9457.EndpointNotMatchedErrorType {
		t.Errorf("Type: got %v, want %v", resp.Type, rfc9457.EndpointNotMatchedErrorType)
	}
	if resp.Title != "Endpoint Not Matched" {
		t.Errorf("Title: got %q, want %q", resp.Title, "Endpoint Not Matched")
	}
	if resp.Status != http.StatusNotFound {
		t.Errorf("Status: got %d, want %d", resp.Status, http.StatusNotFound)
	}
	if resp.Detail == "" {
		t.Error("Detail should not be empty")
	}
	if resp.Instance != req.RequestURI {
		t.Errorf("Instance: got %q, want %q", resp.Instance, req.RequestURI)
	}

	if len(resp.Extensions) != 1 {
		t.Errorf("Number of extensions should be 1, got %d", len(resp.Extensions))
	}
	respExt := resp.Extensions[0]

	ext, ok := respExt.(apiutil.RFC9457Extension)
	if !ok {
		t.Errorf("Extensions is not of type apiutil.RFC9457Extension, got %T instead", respExt)
	}

	// Optional fields should be empty
	if ext.Parameter != "" {
		t.Errorf("Parameter should be empty, got %q", ext.Parameter)
	}
	if ext.ExpectedType != "" {
		t.Errorf("ExpectedType should be empty, got %q", ext.ExpectedType)
	}
	if ext.ReceivedValue != "" {
		t.Errorf("ReceivedValue should be empty, got %q", ext.ReceivedValue)
	}
	if ext.Location != "" {
		t.Errorf("Location should be empty, got %q", ext.Location)
	}
	if ext.Constraint != nil {
		t.Errorf("Constraint should be nil, got %v", ext.Constraint)
	}
	if ext.Suggestion != "" {
		t.Errorf("Suggestion should be empty, got %q", ext.Suggestion)
	}
	if len(ext.ValidationErrors) != 0 {
		t.Errorf("ValidationErrors should be empty, got %d items", len(ext.ValidationErrors))
	}
}

// TestUnprocessableEntityPayload_WithRFC9457 tests UnprocessableEntityPayload with valid RFC9457
func TestUnprocessableEntityPayload_WithRFC9457(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/users/abc", nil)

	inputRFC9457 := &rfc9457.Response{
		Type:     rfc9457.InvalidParameterErrorType,
		Title:    "Invalid Parameter Type",
		Status:   http.StatusUnprocessableEntity,
		Detail:   "Parameter 'id' expected type 'int' but received 'abc'",
		Instance: "/api/users/abc",
		Extensions: []rfc9457.Extension{
			apiutil.RFC9457Extension{
				Parameter:     "id",
				ExpectedType:  "int",
				ReceivedValue: "abc",
				Location:      apiutil.PathLocation,
			},
		},
	}

	payload := apiutil.UnprocessableEntityPayload(req, apiutil.PayloadArgs{
		RFC9457: inputRFC9457,
	})

	// Should return the same Response
	var resp *rfc9457.Response
	if !errors.As(payload, &resp) {
		t.Fatalf("Expected *apiutil.Response, got %T", payload)
	}

	// Should be the exact same instance
	//goland:noinspection GoDirectComparisonOfErrors
	if resp != inputRFC9457 {
		t.Error("UnprocessableEntityPayload should return the same Response instance")
	}

	// Verify it's unchanged
	if resp.Type != rfc9457.InvalidParameterErrorType {
		t.Errorf("Type: got %v, want %v", resp.Type, rfc9457.InvalidParameterErrorType)
	}
	if resp.Title != "Invalid Parameter Type" {
		t.Errorf("Title: got %q, want %q", resp.Title, "Invalid Parameter Type")
	}
	if resp.Status != http.StatusUnprocessableEntity {
		t.Errorf("Status: got %d, want %d", resp.Status, http.StatusUnprocessableEntity)
	}

	if len(resp.Extensions) != 1 {
		t.Errorf("Number of extensions should be 1, got %d", len(resp.Extensions))
	}
	respExt := resp.Extensions[0]

	ext, ok := respExt.(apiutil.RFC9457Extension)
	if !ok {
		t.Errorf("Extensions is not of type apiutil.RFC9457Extension, got %T instead", respExt)
	}

	if ext.Parameter != "id" {
		t.Errorf("Parameter: got %q, want %q", ext.Parameter, "id")
	}
}

// TestUnprocessableEntityPayload_WithNil tests UnprocessableEntityPayload with nil RFC9457
func TestUnprocessableEntityPayload_WithNil(t *testing.T) {
	var resp *rfc9457.Response

	req := httptest.NewRequest("GET", "/api/users/abc", nil)

	payload := apiutil.UnprocessableEntityPayload(req, apiutil.PayloadArgs{})

	// Should return InternalServerError instead
	if !errors.As(payload, &resp) {
		t.Fatalf("Expected *apiutil.Response, got %T", payload)
	}

	// Should be an internal server error (fallback behavior)
	if resp.Type != rfc9457.InternalServerErrorType {
		t.Errorf("Type: got %v, want %v", resp.Type, rfc9457.InternalServerErrorType)
	}
	if resp.Status != http.StatusInternalServerError {
		t.Errorf("Status: got %d, want %d", resp.Status, http.StatusInternalServerError)
	}
	if resp.Title != "Internal Server Error" {
		t.Errorf("Title: got %q, want %q", resp.Title, "Internal Server Error")
	}
}

// TestNewResponsePayload tests the NewResponsePayload function
func TestNewResponsePayload(t *testing.T) {
	content := map[string]interface{}{
		"id":    1,
		"email": "test@example.com",
		"name":  "Test User",
	}

	payload := apiutil.NewResponsePayload(apiutil.ResponsePayloadArgs{
		Content:    content,
		HTTPStatus: http.StatusOK,
		MIMEType:   rfc9457.ApplicationJSON,
	})

	// Verify it implements ResponsePayload
	var resp *rfc9457.Response
	if !errors.As(payload, &resp) {
		t.Fatalf("Expected *apiutil.Response, got %T", payload)
	}

	// Verify HTTPStatusCode
	if payload.HTTPStatusCode() != http.StatusOK {
		t.Errorf("HTTPStatusCode: got %d, want %d", payload.HTTPStatusCode(), http.StatusOK)
	}

	// Verify MIMEType
	if payload.MIMEType() != rfc9457.ApplicationJSON {
		t.Errorf("MIMEType: got %v, want %v", payload.MIMEType(), rfc9457.ApplicationJSON)
	}

	// Verify Content (if ContentGetter interface is available)
	contentGetter, ok := payload.(rfc9457.ContentGetter)
	if ok && contentGetter.Content() == nil {
		t.Error("Content should not be nil")
	}
}

// TestPayloadImplementsInterfaces verifies all payloads implement required interfaces
func TestPayloadImplementsInterfaces(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	testCases := []struct {
		name    string
		payload rfc9457.ResponsePayload
	}{
		{
			name:    "InternalServerErrorPayload",
			payload: apiutil.InternalServerErrorPayload(req, apiutil.PayloadArgs{}),
		},
		{
			name:    "EndpointNotMatchedPayload",
			payload: apiutil.EndpointNotMatchedPayload(req, apiutil.PayloadArgs{}),
		},
		{
			name: "UnprocessableEntityPayload",
			payload: apiutil.UnprocessableEntityPayload(req, apiutil.PayloadArgs{
				RFC9457: &rfc9457.Response{
					Type:   rfc9457.InvalidParameterErrorType,
					Title:  "Test",
					Status: 422,
				},
			}),
		},
		{
			name: "NewResponsePayload",
			payload: apiutil.NewResponsePayload(apiutil.ResponsePayloadArgs{
				Content:    map[string]string{"test": "data"},
				HTTPStatus: 200,
				MIMEType:   rfc9457.ApplicationJSON,
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify ResponsePayload interface
			var resp *rfc9457.Response
			if !errors.As(tc.payload, &resp) {
				t.Fatalf("Expected *apiutil.Response, got %T", tc.payload)
			}

			// Verify HTTPStatusCode method exists
			statusCode := tc.payload.HTTPStatusCode()
			if statusCode < 100 || statusCode > 599 {
				t.Errorf("%s HTTPStatusCode out of valid range: %d", tc.name, statusCode)
			}

			// Verify MIMEType method exists
			mimeType := tc.payload.MIMEType()
			if mimeType == "" {
				t.Errorf("%s MIMEType is empty", tc.name)
			}
		})
	}
}
