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

package compose

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func runGoEnvGOPATH(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", "GOPATH")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("failed go env GOPATH: %w: %s", exitErr, exitErr.Stderr)
		}

		return nil, fmt.Errorf("failed 'go env GOPATH': %w", err)
	}

	return filepath.SplitList(strings.TrimSpace(string(out))), nil
}

type goList struct {
	Dir        string
	ImportPath string
	Module     *struct {
		Path string
		Dir  string
	}
}

func runGoList(ctx context.Context) (goList, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return goList{}, fmt.Errorf("failed go list -json: %w: %s", exitErr, exitErr.Stderr)
		}

		return goList{}, fmt.Errorf("failed go list -json: %w", err)
	}

	var gl goList
	if err := json.Unmarshal(out, &gl); err != nil {
		return goList{}, fmt.Errorf("failed json.Unmarshal: %w", err)
	}

	return gl, nil
}
