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

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/bloomberg/go-testgroup"
)

func Test_dkt_runner(t *testing.T) {
	testgroup.RunSerially(t, &dktRunnerTests{}) // cannot parallelize due to chdir
}

type dktRunnerTests struct{}

func (grp *dktRunnerTests) Version(t *testgroup.T) {
	var stdout, stderr strings.Builder

	debugTraceEnabled := false
	keepExecutable := false
	exitCode := run(debugTraceEnabled, keepExecutable, nil, &stdout, &stderr, "--version")

	t.Zero(exitCode)
	t.Contains(stdout.String(), "dkt runner")
	t.Contains(stdout.String(), "dkt/main")
	t.Contains(stdout.String(), "docker-compose")
}

func (grp *dktRunnerTests) Config(t *testgroup.T) {
	t.Require.NoError(os.Chdir("testdata"))
	defer func() {
		t.NoError(os.Chdir(".."))
	}()

	var stdout strings.Builder

	debugTrace := false
	keepExe := false
	exitCode := run(debugTrace, keepExe, nil, &stdout, nil, "--mode=good", "config")

	t.Zero(exitCode)
	t.Contains(stdout.String(), "version")
}

func (grp *dktRunnerTests) DebugTrace(t *testgroup.T) {
	t.Require.NoError(os.Chdir("testdata"))
	defer func() {
		t.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder

	debugTrace := true
	keepExe := false
	exitCode := run(debugTrace, keepExe, nil, &stdout, &stderr, "--mode=good", "config")

	t.Zero(exitCode)

	t.Contains(stdout.String(), "version")

	t.Contains(stderr.String(), debugPrefix+"dkt runner")
	t.Regexp(debugPrefix+"(current module|not in module-aware mode)", stderr.String())
	t.Regexp(debugPrefix+"(found dkt module|found dkt inside the GOPATH)", stderr.String())
}
