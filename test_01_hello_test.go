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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Test_01_hello(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent test suite in short mode")
	}

	runSuiteWithAndWithoutModules(t, &HelloSuite{
		dir: filepath.Join("testdata", "01_hello"),
	})
}

type HelloSuite struct {
	gopathOrModulesSuite

	dir string
}

func (s *HelloSuite) Test_FailsOutsideDocker() {
	cmd := exec.CommandContext(context.Background(), "go", "test", "-v")
	cmd.Args = append(cmd.Args, coverageArgs(s.T().Name())...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), s.GopathEnvOverride()...)

	out, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("%s", out)
	}
	s.Error(err)
}

func (s *HelloSuite) Test_SucceedsInsideDocker() {
	cmd := exec.CommandContext(context.Background(), "go", "test", "-v")
	cmd.Args = append(cmd.Args, coverageArgs(s.T().Name())...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), s.GopathEnvOverride()...)
	cmd.Env = append(cmd.Env, "DOCKET_MODE=1", "DOCKET_DOWN=1")

	// Since we activated docket, it should succeed inside our docker-compose app.
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)
}
