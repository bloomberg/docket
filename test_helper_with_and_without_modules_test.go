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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

var runningInsideGOPATH = false

func init() {
	// `go list -m` will return false if not inside a module
	cmd := exec.CommandContext(context.Background(), "go", "list", "-m")
	_, err := cmd.CombinedOutput()
	if err != nil {
		runningInsideGOPATH = true
	}
}

type gopathOrModulesSuite struct {
	suite.Suite

	gopath string
}

func (s *gopathOrModulesSuite) SetGOPATH(path string) {
	s.gopath = path
}

func (s *gopathOrModulesSuite) GopathEnvOverride() []string {
	if s.gopath == "" {
		return nil
	}

	return []string{fmt.Sprintf("GOPATH=%s", s.gopath)}
}

type gopathTestingSuite interface {
	suite.TestingSuite

	SetGOPATH(string)
}

func runSuiteWithAndWithoutModules(t *testing.T, s gopathTestingSuite) {
	t.Run("GOPATH mode", func(t *testing.T) {
		if runningInsideGOPATH {
			suite.Run(t, s)
		} else {
			t.Skip("skipping since we are not running inside GOPATH")
		}
	})

	t.Run("module-aware mode", func(t *testing.T) {
		if runningInsideGOPATH {
			// Use a fake GOPATH so that we use module-aware mode.

			relDir, err := ioutil.TempDir(".", "FAKE_GOPATH.")
			if err != nil {
				t.Fatalf("failed to make fake GOPATH: %v", err)
			}
			defer func() {
				if remErr := os.RemoveAll(relDir); remErr != nil {
					t.Fatalf("failed to remove fake GOPATH %q: %v", relDir, remErr)
				}
			}()

			absDir, err := filepath.Abs(relDir)
			if err != nil {
				t.Fatalf("filepath.Abs failed on %q: %v", relDir, err)
			}
			defer func() {
				cmd := exec.CommandContext(context.Background(), "go", "clean", "-modcache")
				cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%s", absDir))
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("failed go clean -modcache: err: %v, out: %s", err, out)
				}
			}()

			s.SetGOPATH(absDir)
			defer func() { s.SetGOPATH("") }()
		}

		suite.Run(t, s)
	})
}
