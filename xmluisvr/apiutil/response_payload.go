package apiutil

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/errutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

type ResponsePayload interface {
	ResponsePayload()
	HTTPStatusCode() int
	MIMEType() rfc9457.MIMEType
	error
}

var responsePayloadArchetype = reflect.TypeOf((*ResponsePayload)(nil))

func GetResponsePayload(pe errutil.ParsedError) (rp ResponsePayload, err error) {
	var ce error
	ce, err = pe.GetCustomError(responsePayloadArchetype)
	if err != nil {
		_ = errors.As(ce, &rp)
	}
	return rp, err
}

func MaybeGetResponsePayload(pe errutil.ParsedError) (rp ResponsePayload) {
	rp, _ = GetResponsePayload(pe)
	return rp
}

var _ ResponsePayload = (*responsePayload)(nil)
var _ rfc9457.ContentGetter = (*responsePayload)(nil)

type responsePayload struct {
	content    any
	httpStatus int
	mimeType   rfc9457.MIMEType
}

type ResponsePayloadArgs struct {
	Content    any
	HTTPStatus int
	MIMEType   rfc9457.MIMEType
}

func NewResponsePayload(args ResponsePayloadArgs) ResponsePayload {
	return &responsePayload{
		content:    args.Content,
		httpStatus: args.HTTPStatus,
		mimeType:   args.MIMEType,
	}
}

func (rp responsePayload) Content() any {
	return rp.content
}

func (responsePayload) ResponsePayload() {}

func (rp responsePayload) HTTPStatusCode() int {
	return rp.httpStatus
}

func (rp responsePayload) MIMEType() rfc9457.MIMEType {
	return rp.mimeType
}
func (rp responsePayload) Error() string {
	return fmt.Sprintf("responsePayload\ncontent=%v\nhttp_status=%d\nmime_type=%v",
		rp.content,
		rp.httpStatus,
		rp.mimeType,
	)
}
