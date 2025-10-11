package apiutil

import (
	"unsafe"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

var paType = (*PayloadArgs)(nil)

type propInfo struct {
	name     string
	zeroFunc func(*PayloadArgs) bool
}

var payloadArgsProps = map[uintptr]propInfo{
	unsafe.Offsetof(paType.Suggestion):        {"Suggestion", func(args *PayloadArgs) bool { return args.Suggestion == "" }},
	unsafe.Offsetof(paType.Location):          {"Location", func(args *PayloadArgs) bool { return args.Location == "" }},
	unsafe.Offsetof(paType.HTTPStatus):        {"HTTPStatus", func(args *PayloadArgs) bool { return args.HTTPStatus == 0 }},
	unsafe.Offsetof(paType.ErrorStyle):        {"ErrorStyle", func(args *PayloadArgs) bool { return args.ErrorStyle == "" }},
	unsafe.Offsetof(paType.Error):             {"Error", func(args *PayloadArgs) bool { return args.Error == nil }},
	unsafe.Offsetof(paType.MissingParameters): {"MissingParameters", func(args *PayloadArgs) bool { return len(args.MissingParameters) == 0 }},
	unsafe.Offsetof(paType.RFC9457):           {"RFC9457", func(args *PayloadArgs) bool { return args.RFC9457 == nil }},
	unsafe.Offsetof(paType.DBQuery):           {"DBQuery", func(args *PayloadArgs) bool { return args.DBQuery == "" }},
	unsafe.Offsetof(paType.EndpointTemplate):  {"EndpointTemplate", func(args *PayloadArgs) bool { return args.EndpointTemplate == "" }},
	unsafe.Offsetof(paType.Detail):            {"Detail", func(args *PayloadArgs) bool { return args.Detail == "" }},
}

type PayloadArgs struct {
	Suggestion        string
	Location          LocationType
	HTTPStatus        int
	ErrorStyle        common.ErrorStyle
	Error             error
	MissingParameters []MissingParameter
	RFC9457           *rfc9457.Response
	DBQuery           dbqvars.QueryString
	EndpointTemplate  string
	Detail            string
	propsUsed         map[uintptr]struct{}
}

func (args *PayloadArgs) clone() *PayloadArgs {
	pa := *args
	pa.propsUsed = make(map[uintptr]struct{})
	if pa.Error == nil {
		pa.Error = ErrUnspecifiedError
	}
	if pa.MissingParameters == nil {
		pa.MissingParameters = make([]MissingParameter, 0)
	}
	return &pa
}

func (args *PayloadArgs) useProp(propId uintptr) {
	args.propsUsed[propId] = struct{}{}
}

// checkUsage checks to see if the developer passed a disallowed non-zero value
// for a PayloadArgs property whose value will not be used by the function OR did
// not pass non-zero PayloadArgs that were use and generates a fatal logging
// error if so. This makes sure the developer is made away that the function will
// not use any of these "disallowed" arg rather than allowing a potentially
// subtle bug to remain in the source code.
// TODO: This uses common.Logger() inside stderrf. See if we can eliminate that import.
func (args *PayloadArgs) checkUsage(rp ResponsePayload) ResponsePayload {
	for propId, info := range payloadArgsProps {
		_, used := args.propsUsed[propId]
		switch {
		case used && info.zeroFunc(args):
			stderrf("\nProperty '%s' was expected TO be passed but had a zero value\n", info.name)
		case !used && info.zeroFunc(args):
			stderrf("\nProperty '%s' was expected to NOT be passed but had a non-zero value\n", info.name)
		}
	}
	return rp
}

func (args *PayloadArgs) GetHTTPStatus() int {
	args.useProp(unsafe.Offsetof(args.HTTPStatus))
	return args.HTTPStatus
}

func (args *PayloadArgs) GetErrorStyle() common.ErrorStyle {
	args.useProp(unsafe.Offsetof(args.ErrorStyle))
	return args.ErrorStyle
}

func (args *PayloadArgs) GetError() error {
	args.useProp(unsafe.Offsetof(args.Error))
	return args.Error
}

func (args *PayloadArgs) GetLocation() LocationType {
	args.useProp(unsafe.Offsetof(args.Location))
	return args.Location
}

func (args *PayloadArgs) GetRFC9457() *rfc9457.Response {
	args.useProp(unsafe.Offsetof(args.RFC9457))
	return args.RFC9457
}

func (args *PayloadArgs) GetSuggestion() string {
	args.useProp(unsafe.Offsetof(args.Suggestion))
	return args.Suggestion
}

func (args *PayloadArgs) GetDetail() string {
	args.useProp(unsafe.Offsetof(args.Detail))
	return args.Detail
}

func (args *PayloadArgs) GetMissingParameters() []MissingParameter {
	args.useProp(unsafe.Offsetof(args.MissingParameters))
	return args.MissingParameters
}

func (args *PayloadArgs) GetDBQuery() dbqvars.QueryString {
	args.useProp(unsafe.Offsetof(args.DBQuery))
	return args.DBQuery
}
