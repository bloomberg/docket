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

	"github.com/stretchr/testify/suite"
)

func Test_dkt_runner(t *testing.T) {
	suite.Run(t, &dktRunnerSuite{})
}

type dktRunnerSuite struct {
	suite.Suite
}

func (s *dktRunnerSuite) Test_version() {
	var stdout, stderr strings.Builder

	debugTraceEnabled := false
	keepExecutable := false
	exitCode := run(debugTraceEnabled, keepExecutable, nil, &stdout, &stderr, "--version")

	s.Zero(exitCode)
	s.Contains(stdout.String(), "dkt runner")
	s.Contains(stdout.String(), "dkt/main")
	s.Contains(stdout.String(), "docker-compose")
}

func (s *dktRunnerSuite) Test_config() {
	s.Require().NoError(os.Chdir("testdata"))
	defer func() {
		s.NoError(os.Chdir(".."))
	}()

	var stdout strings.Builder

	debugTrace := false
	keepExe := false
	exitCode := run(debugTrace, keepExe, nil, &stdout, nil, "--mode=good", "config")

	s.Zero(exitCode)
	s.Contains(stdout.String(), "version")
}

func (s *dktRunnerSuite) Test_debugTrace() {
	s.Require().NoError(os.Chdir("testdata"))
	defer func() {
		s.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder

	debugTrace := true
	keepExe := false
	exitCode := run(debugTrace, keepExe, nil, &stdout, &stderr, "--mode=good", "config")

	s.Zero(exitCode)

	s.Contains(stdout.String(), "version")

	s.Contains(stderr.String(), debugPrefix+"dkt runner")
	s.Regexp(debugPrefix+"(current module|not in module-aware mode)", stderr.String())
	s.Regexp(debugPrefix+"(found dkt module|found dkt inside the GOPATH)", stderr.String())
}
