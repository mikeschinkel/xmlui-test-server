package cfgldr

import (
	"errors"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type APIConfig interface {
	Config()
	IsNil() bool
}

func LoadAPIFileIfExists(apiFile string) (_ APIConfig, err error) {
	var apiV1 *APIDescription
	var apiV2 *APIConfigV2
	// LoadJSON the APIConfig description if provided
	if apiFile == "" {
		goto end
	}
	err = common.CheckFileExists(common.Filepath(apiFile))
	switch {
	case errors.Is(err, common.ErrFileDoesNotExist):
		cliutil.Printf("APIConfig description file %s does not exist", apiFile)
	case errors.Is(err, common.ErrPathIsDir):
		cliutil.Printf("APIConfig description file specified %s is a directory", apiFile)
	case err != nil:
		cliutil.Printf("Unexpected error loading APIConfig description file %s: %v", apiFile, err)
		logger.Error("Error loading APIConfig description file", "api_file", apiFile, "error", err)
	}
	err = nil
	apiV2, err = LoadAPIConfigV2(apiFile)
	if err != nil {
		cliutil.Errorf("Failed to load APIConfig v2 description file: %v", err)
		logger.Error("Error loading APIConfig v2 description file", "api_file", apiFile, "error", err)
		goto end
	}
	if apiV2 != nil {
		goto end
	}
	apiV1, err = LoadAPIDescriptionFromFile(apiFile)
	if err != nil {
		cliutil.Errorf("Failed to load APIConfig v1 description file: %v", err)
		logger.Error("Error loading APIConfig v1 description file", "api_file", apiFile, "error", err)
		goto end
	}
	if apiV1 != nil {
		apiV2 = apiV1.Migrate()
		goto end
	}
	cliutil.Errorf("Failed to load APIConfig description file %s", apiFile)
end:
	if apiV2 != nil {
		cliutil.Printf("APIConfig loaded successfully: %s (v%d)",
			apiV2.Name,
			apiV2.Version,
		)
	}
	return apiV2, err
}
