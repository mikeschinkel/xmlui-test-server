package common

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"log"
	"net/http"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
)

// SendErrorResponse send an error response with the given status code from an HTTP Handler
func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	log.Printf("Error: %s (Status: %d)", message, statusCode)
	http.Error(w, message, statusCode)
}

// SendJSONResponse sends a JSON response with the given status code given an HTTP request
func SendJSONResponse(w http.ResponseWriter, r *http.Request, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Generate JSON response
	responseJSON, err := jsonv2.Marshal(data)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	responseJSON = prettifyJSON(responseJSON)
	// Log the response if enabled - this is the ONLY place where responses should be logged
	maybeEchoResponse(r, responseJSON, statusCode, true)

	// Send the response
	if _, err := w.Write(responseJSON); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Send JSON response with the given status code
func maybeEchoResponse(r *http.Request, responseJSON []byte, statusCode int, verbose bool) {
	// TODO Get Verbose from global Options
	if !verbose {
		goto end
	}

	cliutil.Printf("Request: %s", r.URL.String())
	cliutil.Printf("Status:  %d", statusCode)
	cliutil.Printf("Response:\n%s", string(responseJSON))
end:
	return
}

// Format JSON is a pretty manner
func prettifyJSON(responseJSON []byte) (prettyJSON jsontext.Value) {
	var err error

	prettyJSON = responseJSON
	err = prettyJSON.Indent(jsontext.WithIndent("  "))
	if err != nil {
		cliutil.Errorf("Error prettifying JSON for logging: %v", err)
		cliutil.Errorf("Raw response: %s", string(responseJSON))
		goto end
	}
end:
	return prettyJSON
}
