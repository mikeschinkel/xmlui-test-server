package apiresp

import (
	jsonv2 "encoding/json/v2"
	"log/slog"
	"net/http"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

type Response struct {
	Request    *http.Request
	HTTPWriter http.ResponseWriter
	cliutil.Writer
	*slog.Logger
	cliutil.WriterLogger
}

func NewResponse(args ResponseArgs) *Response {
	return &Response{
		Request:      args.Request,
		HTTPWriter:   args.HTTPWriter,
		Writer:       args.CLIWriter,
		Logger:       args.Logger,
		WriterLogger: cliutil.NewWriterLogger(args.CLIWriter, args.Logger),
	}
}

type ResponseArgs struct {
	HTTPWriter http.ResponseWriter
	Request    *http.Request
	CLIWriter  cliutil.Writer
	Logger     *slog.Logger
}

// Send sends a response to an API requests with the given data and status code
func (r *Response) Send(payload ResponsePayload) {
	var cli cliutil.Writer
	var bytes []byte
	var err error
	var content any
	var getter rfc9457.ContentGetter
	var ok bool
	mimeType := payload.MIMEType()

	w := r.HTTPWriter
	w.Header().Set("Content-Type", string(mimeType))
	w.WriteHeader(payload.HTTPStatusCode())

	internalServerError := func() {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	content = payload
	getter, ok = payload.(rfc9457.ContentGetter)
	if ok {
		content = getter.Content()
	}

	switch mimeType {
	case rfc9457.ApplicationJSON, rfc9457.ApplicationProblemJSON:

		// Generate JSON response
		bytes, err = jsonv2.Marshal(content)
		if err != nil {
			r.WarnError("Error encoding JSON response", "error", err)
			internalServerError()
			goto end
		}
		bytes = prettifyJSON(bytes)
	case "":
		r.WarnError("No response MIME type specified", "http_request", r.Request)
		internalServerError()
		goto end
	default:
		r.WarnError("Unsupported response MIME type", "mime_type", mimeType, "http_request", r.Request)
		internalServerError()
		goto end
	}

	cli = r.Writer.V2()
	cli.Printf("Request: %s\n", r.Request.URL.String())
	cli.Printf("Status:  %d\n", payload.HTTPStatusCode())
	cli.V3().Printf("Response:\n%s", string(bytes))

	// Send the response
	_, err = r.HTTPWriter.Write(bytes)
	if err != nil {
		r.WarnError("Error writing response", "http_request", r.Request, "error", err)
	}
end:
	return
}
