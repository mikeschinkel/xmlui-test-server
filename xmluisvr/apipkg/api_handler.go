package apipkg

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiresp"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/errparsr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

// HandleAPIFunc returns an HTTP handler function that processes API requests.
// The handler:
//  1. Matches the request path against configured endpoints
//  2. Extracts path parameters and request body
//  3. Loads the SQL query for the matched endpoint
//  4. Executes the query against the database
//  5. Returns the results as JSON
//
// Returns 404 for unmatched routes and 500 for server errors.
func (api *API) HandleAPIFunc(ctx Context, db dbpkg.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var result pathvars.MatchResult
		var err error
		var args HandlerHelperArgs

		api.Writer.Printf("Handling API Request: %s %s\n", r.Method, r.URL.Path)

		args = HandlerHelperArgs{
			HTTPRequest: r,
			MatchResult: result,
			Database:    db,
			APIResponse: apiresp.NewResponse(apiresp.ResponseArgs{
				HTTPWriter: w,
				Request:    r,
				CLIWriter:  api.Writer,
				Logger:     api.Logger,
			}),
		}

		// Find the matching endpoint
		args.Endpoint, args.MatchResult, err = api.getMatchingEndpoint(args)
		if err != nil {
			goto end
		}

		// Now get the query values
		args.QueryValues, err = api.getQueryValues(args)
		if err != nil {
			goto end
		}

		// Now call the SQL query
		args.QueryResult, err = api.GetQueryResult(ctx, args)
		if err != nil {
			goto end
		}

		// Now get the response content
		args.Content, err = api.GetResponseContent(args)
		if err != nil {
			goto end
		}

		// Finally, send the success response. `args.Content` is expected to
		// contain the value to return as JSON.
		api.SendSuccessResponse(args.SendResponseArgs(nil))
	end:
		if err != nil {
			// Or, send the error if it was an error This assumes that err will have been
			// joined with a `ResponsePayload` — which by declaration is also an `error` —
			// one example being rfc9457.Response.
			api.SendErrorResponse(args.SendResponseArgs(err))
		}
		return
	}
}

func (api *API) getQueryValues(args HandlerHelperArgs) (queryValues []any, err error) {
	var missing []apiresp.MissingParameter

	r := args.HTTPRequest
	result := args.MatchResult
	endpoint := args.Endpoint

	queryValues, missing, err = endpoint.GetParameterValues(ParameterValuesSource{
		ValuesMap:  result.ValuesMap(),
		BodyReader: r.Body,
		Headers:    nil, // TODO: Not yet supported
	})
	if err != nil {
		if len(missing) != 0 {
			err = errors.Join(
				ErrQueryValuesExtractionFailed,
				apiresp.MissingParametersPayload(r, apiresp.PayloadArgs{
					MissingParameters: missing,
				}),
				err,
			)
			goto end
		}
		err = errors.Join(
			ErrQueryValuesExtractionFailed,
			apiresp.CurrentlyUnhandledErrorPayload(r, apiresp.PayloadArgs{
				Location: "CHANGE ME", // TODO: Determine appropriate value by breakpoint debugging during tests
				Error:    err,
			}),
			err,
		)
	}
end:
	return queryValues, err
}

func (api *API) GetQueryResult(ctx Context, args HandlerHelperArgs) (dbResult apiresp.QueryResult, err error) {
	var rows dbpkg.QueryResult
	endpoint := args.Endpoint
	dbq := endpoint.ParsedQuery
	qs := dbq.QueryString()
	rows, err = dbpkg.ExecuteQuery(ctx, args.Database, qs, args.QueryValues)
	if err != nil {
		err = errors.Join(
			ErrQueryValuesExtractionFailed,
			apiresp.QueryFailedPayload(args.HTTPRequest, apiresp.PayloadArgs{
				ErrorStyle: api.Options.ErrorStyle,
			}),
			err,
		)
		goto end
	}
	dbResult = apiresp.NewQueryResult(rows)

	api.V3().InfoPrint("Database query submitted.",
		"requestor_ip", args.HTTPRequest.RemoteAddr,
		"query", strings.Replace(string(args.DBQuery), "\n", " ", -1),
	)

end:
	return dbResult, err
}

func (api *API) getMatchingEndpoint(args HandlerHelperArgs) (ep *Endpoint, mr pathvars.MatchResult, err error) {
	// Find the matching endpoint
	mr, err = api.tryMatchingRequest(args)
	if err != nil {
		goto end
	}
	// Assign the endpoint to the helper args
	ep = api.Endpoints[mr.Index]
end:
	return ep, mr, err
}

func (api *API) handleFailedMatch(result pathvars.MatchResult, err error, args HandlerHelperArgs) error {
	var httpStatus int
	var resp *rfc9457.Response
	var pe errparsr.ParsedError
	var hasErrors bool

	r := args.HTTPRequest

	pe, _ = errparsr.ParseError(err)
	err = pe.MaybeGetCustomError(rfc9457.ResponseArchetype)
	if err != nil && errors.As(err, &resp) {
		httpStatus = resp.Status
	}
	if httpStatus == 0 {
		httpStatus = pe.MaybeGetIntDetail("http_status")
	}
	hasErrors = pe.HasErrors()

	// TODO There are probably more cases we need to add
	switch {
	case httpStatus == 0 && !hasErrors:
		goto end
	case httpStatus == http.StatusUnprocessableEntity:
		err = errors.Join(
			ErrRouteMatchingFailed,
			apiresp.UnprocessableEntityPayload(r, apiresp.PayloadArgs{
				RFC9457: resp,
			}),
			err,
		)
		goto end
	case hasErrors && httpStatus != 0:
		err = errors.Join(
			ErrRouteMatchingFailed,
			// TODO Verify that "matching_request" is appropriate for "location"
			apiresp.CurrentlyUnhandledErrorPayload(r, apiresp.PayloadArgs{
				Location:   "CHANGE ME", // TODO: Use debugging to identify what values are useful here
				HTTPStatus: httpStatus,
				Error:      err,
			}),
			err,
		)
	default:
		err = errors.Join(
			ErrRouteMatchingFailed,
			apiresp.InternalServerErrorPayload(r, apiresp.PayloadArgs{}),
			err,
		)
		goto end
	}
end:
	return err
}

func (api *API) tryMatchingRequest(args HandlerHelperArgs) (result pathvars.MatchResult, err error) {
	var pve pathvars.ParameterValidationError

	r := args.HTTPRequest

	result, err = api.Router.Match(r)
	if errors.As(err, &pve) {
		err = errors.Join(
			pathvars.ErrInvalidParameter,
			pve.Err,
			apiresp.InvalidURLParameterErrorPayload(r, apiresp.PayloadArgs{PVE: &pve}),
			err,
		)
		goto end
	}
	if errors.Is(err, pathvars.ErrNoMatch) {
		err = errors.Join(
			ErrRouteNotMatched,
			apiresp.EndpointNotMatchedPayload(r, apiresp.PayloadArgs{}),
			err,
		)
		goto end
	}
	if err != nil {
		err = api.handleFailedMatch(result, err, args)
	}
end:
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("http_method=%s", r.Method),
			fmt.Errorf("url_path=%s", r.URL.Path),
		)
	}
	return result, err
}

type SendResponseArgs struct {
	Content     any
	HTTPRequest *http.Request
	APIResponse *apiresp.Response
	Error       error
}

func (api *API) SendSuccessResponse(args SendResponseArgs) {
	// TODO Validate result types and column types
	//      OR MAYBE THAT IS ALREADY BEING HANDLED UPSTREAM?
	args.APIResponse.Send(apiresp.NewResponsePayload(apiresp.ResponsePayloadArgs{
		Content:    args.Content,
		HTTPStatus: http.StatusOK,
		MIMEType:   rfc9457.ApplicationJSON,
	}))
}

func (api *API) SendErrorResponse(args SendResponseArgs) {
	// If an error occurred it would be encoded as a response payload so we parse the
	// error, extract the payload, and return as the response.
	pe, _ := errparsr.ParseError(args.Error)
	rp := apiresp.MaybeGetResponsePayload(pe)
	if rp == nil {
		// Okay to ignore error as CurrentlyUnhandledErrorPayload() will report invalid location
		location, _ := pe.GetDetail("location")
		rp = apiresp.CurrentlyUnhandledErrorPayload(args.HTTPRequest, apiresp.PayloadArgs{
			Error:    errors.Join(ErrNoResponsePayloadFound, args.Error),
			Location: apiresp.LocationType(location),
		})
	}
	args.APIResponse.Send(rp)

}
