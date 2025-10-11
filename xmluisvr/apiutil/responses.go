package apiutil

import (
	"encoding/json/jsontext"
	"log"
	"log/slog"
	"net/http"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
)

// SendErrorResponse send an error response with the given status code from an HTTP Handler
func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	log.Printf("Error: %s (Status: %d)", message, statusCode)
	http.Error(w, message, statusCode)
}

// SendJSONResponse sends a JSON response with the given status code given an HTTP request
func SendJSONResponse(w http.ResponseWriter, r *http.Request, data ResponsePayload, statusCode int) {

	response := NewResponse(ResponseArgs{
		HTTPWriter: w,
		Request:    r,
		CLIWriter:  cliutil.GetWriter(),
		Logger:     slog.Default(), // TODO This function will soon be remove so this default logger is just temporary
	})
	response.Send(data)
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
