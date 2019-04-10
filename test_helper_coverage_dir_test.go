// Copyright 2019 Bloomberg Finance L.P.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
