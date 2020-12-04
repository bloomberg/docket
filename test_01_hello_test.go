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

package docket_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bloomberg/go-testgroup"
)

func Test_01_hello(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent tests in short mode")
	}

	testgroup.RunSerially(t, &HelloTests{
		dir: filepath.Join("testdata", "01_hello"),
	})
}

type HelloTests struct {
	dir string
}

func (grp *HelloTests) FailsOutsideDocker(t *testgroup.T) {
	cmd := exec.Command("go", "test", "-v")
	cmd.Args = append(cmd.Args, goTestCoverageArgs(t.Name())...)
	cmd.Args = append(cmd.Args, goTestRaceDetectorArgs()...)
	cmd.Dir = grp.dir

	out, err := cmd.CombinedOutput()
	t.Errorf(err, "output: %q", out)
}

func (grp *HelloTests) SucceedsInsideDocker(t *testgroup.T) {
	cmd := exec.Command("go", "test", "-v")
	cmd.Args = append(cmd.Args, goTestCoverageArgs(t.Name())...)
	cmd.Args = append(cmd.Args, goTestRaceDetectorArgs()...)
	cmd.Dir = grp.dir
	cmd.Env = append(os.Environ(), "DOCKET_MODE=1", "DOCKET_DOWN=1")

	// Since we activated docket, it should succeed inside our docker-compose app.
	out, err := cmd.CombinedOutput()
	t.NoError(err, "output: %q", out)
}
