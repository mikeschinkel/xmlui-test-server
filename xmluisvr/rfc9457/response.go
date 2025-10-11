package rfc9457

import (
	"encoding/json"
	"net/http"
	"reflect"
)

var _ ResponsePayload = (*Response)(nil)
var _ error = (*Response)(nil)

var ResponseArchetype = reflect.TypeOf((*Response)(nil))

type Extension interface{}

type Response struct {
	Type       ErrorTypeURI `json:"type"`
	Title      string       `json:"title"`
	Status     int          `json:"status"`
	Detail     string       `json:"detail,omitempty"`
	Instance   string       `json:"instance,omitempty"`
	Extensions []Extension  `json:"extensions,omitempty"`
}

func (r *Response) AddExtension(ext Extension) {
	r.Extensions = append(r.Extensions, ext)
}

func (r *Response) Error() string {
	return SprintfMany(r.Title, "\n",
		"title=%s", r.Title,
		"error_type=%s", r.Type,
		"error_detail=%s", r.Detail,
		"http_status=%d", r.Status,
		"instance=%s", r.Instance,
	)
}

func (r *Response) MIMEType() MIMEType {
	return ApplicationProblemJSON
}

func (r *Response) HTTPStatusCode() int {
	return r.Status
}

func (*Response) ResponsePayload() {}

func NewResponse(args ResponseArgs) *Response {
	return &Response{
		Type:       args.Type,
		Title:      args.Title,
		Status:     args.Status,
		Detail:     args.Detail,
		Instance:   args.Instance,
		Extensions: args.Extensions,
	}
}

type ResponseArgs struct {
	Type       ErrorTypeURI `json:"type"`
	Title      string       `json:"title"`
	Status     int          `json:"status"`
	Detail     string       `json:"detail"`
	Instance   string       `json:"instance"`
	Extensions []Extension  `json:"extensions"`
}

func (r *ResponseArgs) AddExtension(ext Extension) {
	r.Extensions = append(r.Extensions, ext)
}

func (r *Response) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/problem+json") // RFC 9457 media type
	w.WriteHeader(r.Status)
	return json.NewEncoder(w).Encode(r)
}
