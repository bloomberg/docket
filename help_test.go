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
	"bytes"
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
)

func Test_help(t *testing.T) {
	suite.Run(t, &HelpSuite{})
}

type HelpSuite struct {
	suite.Suite
}

func (s *HelpSuite) Test_runGoTest() {
	cmd := exec.CommandContext(context.Background(), "go", "test", "-help-docket")
	cmd.Args = append(cmd.Args, goTestCoverageArgs(s.T().Name())...)
	cmd.Args = append(cmd.Args, goTestRaceDetectorArgs()...)

	// When run inside go test,
	//   All test output and summary lines are printed to the go command's
	//   standard output, even if the test printed them to its own standard
	//   error. (The go command's standard error is reserved for printing
	//   errors building the tests.)

	out, err := cmd.CombinedOutput()

	s.Error(err)
	s.Regexp("Help for using docket:", string(out))
}

func (s *HelpSuite) Test_writeHelp() {
	var buf bytes.Buffer
	writeHelp(&buf)
	s.Regexp("Help for using docket:", buf.String())
}
