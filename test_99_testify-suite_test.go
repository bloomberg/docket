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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"
)

func Test_99_testify_suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent test suite in short mode")
	}

	suite.Run(t, &TestifySuiteSuite{
		Suite: suite.Suite{},
		dir:   filepath.Join("testdata", "99_testify-suite"),
	})
}

type TestifySuiteSuite struct {
	suite.Suite

	dir string
}

func (s *TestifySuiteSuite) Test_All() {
	s.runGoTest()
}

func (s *TestifySuiteSuite) Test_SuiteLevel_OnlySubtestA() {
	output, sawA, sawB, sawC, sawOthers := s.testSubtestA(true)

	s.Equalf(true, sawA, "should have seen TestA, output: %s", output)
	s.Equalf(false, sawB, "should not have seen TestB, output: %s", output)
	s.Equalf(false, sawC, "should not have seen TestC, output: %s", output)
	s.Equalf(false, sawOthers, "should not have seen other tests, output: %s", output)
}

func (s *TestifySuiteSuite) Test_SuiteLevel_EverythingButSubtestA() {
	output, sawA, sawB, sawC, sawOthers := s.testSubtestA(false)

	s.Equalf(false, sawA, "should not have seen TestA, output: %s", output)
	s.Equalf(true, sawB, "should have seen TestB, output: %s", output)
	s.Equalf(true, sawC, "should have seen TestC, output: %s", output)
	s.Equalf(false, sawOthers, "should not have seen other tests, output: %s", output)
}

//------------------------------------------------------------------------------

func (s *TestifySuiteSuite) runGoTest(arg ...string) []byte {
	cmd := exec.Command("go", "test", "-v")
	cmd.Args = append(cmd.Args, goTestCoverageArgs(s.T().Name())...)
	cmd.Args = append(cmd.Args, goTestRaceDetectorArgs()...)
	cmd.Args = append(cmd.Args, arg...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), "DOCKET_MODE=full", "DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)

	return out
}

// Helper routine that either runs ONLY subtestA or everything EXCEPT subtestA.
func (s *TestifySuiteSuite) testSubtestA(includeA bool) (
	output []byte, sawA, sawB, sawC, sawOthers bool,
) {
	negation := ""
	if !includeA {
		negation = "^"
	}
	runArg := fmt.Sprintf("-run=DocketRunAtSuiteLevel/Test[%sA]", negation)

	output = s.runGoTest(runArg)

	ranTest := regexp.MustCompile(`^=== RUN   Test.+/Test[A-Z]$`)

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		txt := scanner.Text()
		if txt == "PASS" {
			break
		}
		if ranTest.MatchString(txt) {
			switch txt[len(txt)-1] {
			case 'A':
				sawA = true
			case 'B':
				sawB = true
			case 'C':
				sawC = true
			default:
				sawOthers = true
			}
		}
	}

	return output, sawA, sawB, sawC, sawOthers
}
