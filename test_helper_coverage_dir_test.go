package docket

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var coverageDir = os.Getenv("COVERAGE_DIR")

func init() {
	if coverageDir == "" {
		return
	}

	if err := os.MkdirAll(coverageDir, 0755); err != nil {
		panic(fmt.Sprintf("could not mkdir %q: %v", coverageDir, err))
	}
}

func coverageArgs(testName string) []string {
	if coverageDir == "" {
		return nil
	}

	relPath := filepath.Join(coverageDir, testName)
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(filepath.Dir(absPath), 0755)

	return []string{
		"-coverprofile", absPath,
		"-coverpkg", strings.Join([]string{
			"github.com/bloomberg/docket",
			"github.com/bloomberg/docket/internal/...",
		}, ","),
	}
}
