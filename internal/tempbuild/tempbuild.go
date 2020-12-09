// Copyright 2020 Bloomberg Finance L.P.
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

package tempbuild

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

// Build a package at a temporary location.
//
// Instead of installing a program in a global location like GOBIN, you can use Build
// to make a temporary/private copy of the program.
func Build(ctx context.Context, packageSpec, tempFilePattern string) (string, error) {
	file, err := ioutil.TempFile("", tempFilePattern)
	if err != nil {
		return "", fmt.Errorf("failed ioutil.TempFile: %w", err)
	}

	path := file.Name()
	file.Close()

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", path, packageSpec)
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		os.Remove(path)

		return "", fmt.Errorf("go build failed: %w: %s", err, buildOutput)
	}

	return path, nil
}
