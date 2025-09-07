package cfgldr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Context = context.Context

var _ Config = (*APIDescription)(nil)

// APIDescription is deprecated; use APIConfigV2
type APIDescription struct {
	APIVersion  string               `json:"apiVersion"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	BasePath    string               `json:"basePath"`
	Endpoints   []EndpointDefinition `json:"endpoints"`
	pathRegexps map[string]*regexp.Regexp
}

func (d *APIDescription) Migrate() *APIConfigV2 {
	endpoints := make([]*APIEndpointV2, 0, len(d.Endpoints))
	for _, ep := range d.Endpoints {
		endpoints = append(endpoints, ep.Migrate()...)
	}
	return &APIConfigV2{
		SchemaVersion: APIConfigV2SchemaVersion,
		Name:          d.Description,
		BasePath:      d.BasePath,
		Webroot:       ".",
		Endpoints:     endpoints,
		SourceFile:    "",
	}
}

func (*APIDescription) Config() {}

// EndpointDefinition models an API endpoint in v1 api.json schema
// Deprecated — use APIEndpoint instead
type EndpointDefinition struct {
	Path    string                      `json:"path"`
	Methods map[string]MethodDefinition `json:"methods"`
}

// Migrate migrates an EndpointDefinition to an APIEndpointV2
func (d *EndpointDefinition) Migrate() (eps []*APIEndpointV2) {
	eps = make([]*APIEndpointV2, 0, len(d.Methods))
	for name, obj := range d.Methods {
		params := make(map[string]string, len(d.Methods))
		for _, p := range obj.Params {
			params[p] = "any"
		}
		name = strings.ToUpper(name)
		eps = append(eps, &APIEndpointV2{
			Endpoint:     fmt.Sprintf("%s %s", name, d.Path),
			Description:  obj.Description,
			Query:        obj.SQL,
			QueryFile:    obj.SQLFile,
			Params:       params,
			RowsExpected: "many?", // TODO: Move the constants to common?
			RowType:      "any",
			method:       name,
			path:         d.Path,
		})
	}
	return eps
}

// MethodDefinition models an API endpoint method in v1 api.json schema
// Deprecated — use APIEndpoint instead
type MethodDefinition struct {
	Description string   `json:"description"`
	SQL         string   `json:"sql,omitempty"`
	SQLFile     string   `json:"sqlFile,omitempty"`
	Params      []string `json:"params,omitempty"`
}

func LoadAPIDescriptionFromFile(file string) (d *APIDescription, err error) {
	var data []byte
	data, err = os.ReadFile(file)
	if errors.Is(os.ErrNotExist, err) {
		goto end
	}
	if err != nil {
		err = errors.Join(ErrReadFailed, err)
		goto end
	}
	d = &APIDescription{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		d = nil
		err = errors.Join(ErrParseFailed, err)
		goto end
	}

end:
	return d, err
}
