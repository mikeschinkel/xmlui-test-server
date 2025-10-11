package apipkg

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

type HandlerHelperArgs struct {
	HTTPRequest *http.Request
	MatchResult pathvars.MatchResult
	APIResponse *apiutil.Response
	Database    dbpkg.Database
	Endpoint    *Endpoint
	QueryValues []any
	QueryResult apiutil.QueryResult
	DBQuery     dbqvars.QueryString
	RequestBody bytes.Buffer
	Content     any
	TargetURL   *url.URL
	URLPath     common.URLPath
}

func (args HandlerHelperArgs) GetQueryString() (qs dbqvars.QueryString, err error) {
	var dbq dbqvars.ParsedQuery
	if args.DBQuery != "" {
		qs = args.DBQuery
	}
	if args.Endpoint == nil {
		err = apiutil.ErrNeitherDBQueryNorEndpointSet
		goto end
	}
	dbq = args.Endpoint.ParsedQuery
	if dbq == nil {
		err = apiutil.ErrNeitherDBQueryNorEndpointSet
		goto end
	}
	qs = dbq.QueryString()
end:
	return qs, err
}

func (args HandlerHelperArgs) APIEndpointRequested() string {
	r := args.HTTPRequest
	return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
}

func (args HandlerHelperArgs) SendResponseArgs(err error) SendResponseArgs {
	return SendResponseArgs{
		HTTPRequest: args.HTTPRequest,
		APIResponse: args.APIResponse,
		Content:     args.Content,
		Error:       err,
	}
}

func (args HandlerHelperArgs) ErrorMeta(err error) []any {
	dbq := args.Endpoint.ParsedQuery
	meta := []any{
		"endpoint", args.Endpoint.Endpoint(),
		"url_params", args.MatchResult.ValuesMap(),
		"sql_query", dbq.QueryString(),
		"sql_params", dbq.Parameters(),
		// TODO Add more properties?
	}
	if err != nil {
		meta = append(meta, err)
	}
	return meta
}

func (api *API) GetResponseContent(args HandlerHelperArgs) (content any, err error) {
	// Now get the response content
	result := args.MatchResult
	dbResult := args.QueryResult

	content, err = args.QueryResult.GetByCardinality(dbqvars.Cardinality(result.Route.Cardinality))
	switch {
	case errors.Is(err, dbpkg.ErrManyRowsExpectedZeroReturned):
		fallthrough

	case errors.Is(err, dbpkg.ErrOneRowExpectedZeroReturned):
		err = errors.Join(
			ErrQueryValuesExtractionFailed,
			apiutil.NoResultsPayload(args.HTTPRequest, apiutil.PayloadArgs{
				Error:            err,
				ErrorStyle:       api.Options.ErrorStyle,
				DBQuery:          args.Endpoint.ParsedQuery.QueryString(),
				EndpointTemplate: result.Route.Endpoint(),
			}),
			err,
		)
		goto end

	case errors.Is(err, dbpkg.ErrOneRowExpectedManyReturned):
		err = errors.Join(
			ErrQueryValuesExtractionFailed,
			fmt.Errorf("rows_returned=%d", len(dbResult)),
			apiutil.NoResultsPayload(args.HTTPRequest, apiutil.PayloadArgs{
				Error:            err,
				ErrorStyle:       api.Options.ErrorStyle,
				DBQuery:          args.Endpoint.ParsedQuery.QueryString(),
				EndpointTemplate: result.Route.Endpoint(),
			}),
			err,
		)

	case err != nil:
		err = errors.Join(
			ErrQueryValuesExtractionFailed,
			apiutil.CurrentlyUnhandledErrorPayload(args.HTTPRequest, apiutil.PayloadArgs{
				Location:         "CHANGE ME", // TODO: Determine appropriate value by breakpoint debugging during tests
				Error:            err,
				EndpointTemplate: result.Route.Endpoint(),
			}),
			err,
		)

	default:
		// S'all good, man!

	}
	if err != nil {
		err = errors.Join(ErrGettingResponseContent, err)
	}
end:
	return content, err
}
