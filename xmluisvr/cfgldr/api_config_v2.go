package cfgldr

import (
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

const (
	APIConfigV2Version = 2
	APIConfigV2Schema  = "https://schemas.xmlui.org/v2/test-server/api-schema.json"
)

type APIConfig interface {
	Config()
	IsNil() bool
}

var _ APIConfig = (*APIConfigV2)(nil)

type APIConfigV2 struct {
	Schema     string           `json:"$schema"`
	Version    int              `json:"version"`
	Name       string           `json:"name"`
	BasePath   string           `json:"base_path"`
	Webroot    string           `json:"webroot"`
	Endpoints  []*APIEndpointV2 `json:"endpoints"`
	SourceFile string           `json:"-"`
}

func (c *APIConfigV2) IsNil() (isNil bool) {
	var v reflect.Value

	isNil = true
	if c == nil {
		goto end
	}
	v = reflect.ValueOf(c)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		isNil = v.IsNil()
	default:
		isNil = false
	}
end:
	return isNil
}

func NewAPIConfigV2(webroot string) *APIConfigV2 {
	return &APIConfigV2{
		Schema:     APIConfigV2Schema,
		Version:    APIConfigV2Version,
		Name:       fmt.Sprintf("User-definable %s APIConfig", common.AppName),
		BasePath:   "/api",
		Webroot:    webroot,
		Endpoints:  make([]*APIEndpointV2, 0),
		SourceFile: "./api.json",
	}
}

func (*APIConfigV2) Config() {}

func (c *APIConfigV2) normalizeEndpoints() {
	if c.Endpoints == nil {
		c.Endpoints = make([]*APIEndpointV2, 0)
	}
	if len(c.Endpoints) == 0 {
		goto end
	}
	for _, ep := range c.Endpoints {
		ep.Normalize()
	}
end:
	return
}

func (c *APIConfigV2) Normalize(sourceFile string) {
	c.SourceFile = sourceFile
	if c.Schema == "" {
		c.Schema = APIConfigV2Schema
	}
	if c.Version == 0 {
		c.Version = APIConfigV2Version
	}
	if c.BasePath == "" {
		c.BasePath = DefaultAPIBasePath
	}
	if c.Webroot == "" {
		c.Webroot = DefaultWebroot
	}
	c.normalizeEndpoints()
}

func (c *APIConfigV2) AddEndpoint(endpoint *APIEndpointV2) {
	c.Endpoints = append(c.Endpoints, endpoint)
}

func (c *APIConfigV2) Migrate(oldCfg Config) (newCfg *APIConfigV2) {
	v1 := oldCfg.(*APIDescription)
	common.Noop(v1) // TODO Implement migration
	return new(APIConfigV2)
}

func LoadAPIConfigV2(apiFile string) (c *APIConfigV2, err error) {
	var data []byte
	if apiFile == "" {
		goto end
	}
	data, err = os.ReadFile(string(apiFile))
	if errors.Is(os.ErrNotExist, err) {
		err = fmt.Errorf("invalid APIConfig description file: %w", err)
		goto end
	}
	if err != nil {
		err = errors.Join(ErrReadFailed, err)
		goto end
	}
	c = new(APIConfigV2)
	err = jsonv2.Unmarshal(data, c)
	if err != nil {
		c = nil
		// TODO: Provide user better feedback as to what actually failed.
		err = errors.Join(ErrParseFailed, err)
		goto end
	}
end:
	return c, err
}
