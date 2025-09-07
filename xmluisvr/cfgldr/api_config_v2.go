package cfgldr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/xmlui-org/xmluisvr/common"
)

const (
	APIConfigV2SchemaVersion = 2
	APIConfigV2Schema        = "https://schemas.xmlui.org/v2/test-server-api-schema.json"
)

type APIConfig interface {
	Config()
}
type APIConfigV2 struct {
	Schema        string           `json:"$schema"`
	SchemaVersion int              `json:"$schemaVersion"`
	Name          string           `json:"name"`
	BasePath      string           `json:"base_path"`
	Webroot       string           `json:"webroot"`
	Endpoints     []*APIEndpointV2 `json:"endpoints"`
	SourceFile    string           `json:"-"`
}

func (c *APIConfigV2) AddEndpoint(endpoint *APIEndpointV2) {
	c.Endpoints = append(c.Endpoints, endpoint)
}

func NewAPIConfigV2(webroot string) *APIConfigV2 {
	return &APIConfigV2{
		Schema:        APIConfigV2Schema,
		SchemaVersion: APIConfigV2SchemaVersion,
		Name:          fmt.Sprintf("User-definable %s API", common.AppName),
		BasePath:      "/api",
		Webroot:       webroot,
		Endpoints:     make([]*APIEndpointV2, 0),
		SourceFile:    "./api.json",
	}
}

func (*APIConfigV2) Config() {}

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
	data, err = os.ReadFile(apiFile)
	if errors.Is(os.ErrNotExist, err) {
		err = fmt.Errorf("invalid API description file: %w", err)
		goto end
	}
	if err != nil {
		err = errors.Join(ErrReadFailed, err)
		goto end
	}
	c = new(APIConfigV2)
	err = json.Unmarshal(data, c)
	if err != nil {
		c = nil
		// TODO: Provide user better feedback as to what actually failed.
		err = errors.Join(ErrParseFailed, err)
		goto end
	}
end:
	return c, err
}
