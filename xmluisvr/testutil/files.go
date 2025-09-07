package testutil

import (
	"os"
	"testing"
)

func LoadFile(t *testing.T, file string, mustLoad bool) (data []byte) {
	var err error

	data, err = os.ReadFile(file)
	if err != nil && mustLoad {
		t.Fatal(err)
	}

	return data
}
