package cfgldr

import (
	"errors"

	"github.com/xmlui-org/xmluisvr/cliutil"
	"github.com/xmlui-org/xmluisvr/common"
)

func LoadAPIFile(apiFile string) (_ APIConfig, err error) {
	var apiV1 *APIDescription
	var apiV2 *APIConfigV2
	// LoadJSON the API description if provided
	if apiFile == "" {
		goto end
	}
	err = common.CheckFileExists(apiFile)
	switch {
	case errors.Is(err, common.ErrFileDoesNotExist):
		cliutil.Printf("API description file %s does not exist", apiFile)
	case errors.Is(err, common.ErrPathIsDir):
		cliutil.Printf("API description file specified %s is a directory", apiFile)
	case err != nil:
		cliutil.Printf("Unexpected error loading API description file %s: %v", apiFile, err)
		logger.Error("Error loading API description file", "api_file", apiFile, "error", err)
	}
	err = nil
	apiV2, err = LoadAPIConfigV2(apiFile)
	if err != nil {
		cliutil.Errorf("Failed to load API v2 description file: %v", err)
		logger.Error("Error loading API v2 description file", "api_file", apiFile, "error", err)
		goto end
	}
	if apiV2 != nil {
		goto end
	}
	apiV1, err = LoadAPIDescriptionFromFile(apiFile)
	if err != nil {
		cliutil.Errorf("Failed to load API v1 description file: %v", err)
		logger.Error("Error loading API v1 description file", "api_file", apiFile, "error", err)
		goto end
	}
	if apiV1 != nil {
		apiV2 = apiV1.Migrate()
		goto end
	}
	cliutil.Errorf("Failed to load API description file %s", apiFile)
end:
	if apiV2 != nil {
		cliutil.Printf("API loaded successfully: %s (v%d)",
			apiV2.Name,
			apiV2.SchemaVersion,
		)
	}
	return apiV2, err
}
