package fsutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteJSONFile(file string, value any, filePerms, dirPerms os.FileMode) (err error) {
	var bytes []byte
	err = os.MkdirAll(filepath.Dir(file), dirPerms)
	if err != nil {
		err = fmt.Errorf("failed to create directory %s", filepath.Dir(file))
		goto end
	}

	bytes, err = json.MarshalIndent(value, "", "\t")
	if err != nil {
		err = fmt.Errorf("failed to marshal value of type '%T' to json", value)
		goto end
	}

	err = os.WriteFile(file, []byte(bytes), filePerms)
	if err != nil {
		err = fmt.Errorf("failed to write test file; %s", file)
		goto end
	}

end:
	return err
}
