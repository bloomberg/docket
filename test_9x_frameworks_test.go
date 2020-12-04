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

	"github.com/bloomberg/go-testgroup"
)

func Test_9x_frameworks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent test group in short mode")
	}

	testgroup.RunSerially(t, &frameworks{})
}

type frameworks struct{}

func (grp *frameworks) Testgroup(t *testgroup.T) {
	t.RunSerially(&frameworkTests{dir: filepath.Join("testdata", "98_testgroup")})
}

func (grp *frameworks) TestifySuite(t *testgroup.T) {
	t.RunSerially(&frameworkTests{dir: filepath.Join("testdata", "99_testify-suite")})
}

//------------------------------------------------------------------------------

type frameworkTests struct {
	dir string
}

func (grp *frameworkTests) All(t *testgroup.T) {
	grp.runGoTest(t)
}

func (grp *frameworkTests) Subset(t *testgroup.T) {
	t.RunSerially(&frameworkSubsetTests{parent: grp})
}

func (grp *frameworkTests) runGoTest(t *testgroup.T, arg ...string) []byte {
	cmd := exec.Command("go", "test", "-v")
	cmd.Args = append(cmd.Args, goTestCoverageArgs(t.Name())...)
	cmd.Args = append(cmd.Args, goTestRaceDetectorArgs()...)
	cmd.Args = append(cmd.Args, arg...)
	cmd.Dir = grp.dir
	cmd.Env = append(os.Environ(), "DOCKET_MODE=full", "DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	t.NoError(err)

	return out
}

//------------------------------------------------------------------------------

type frameworkSubsetTests struct {
	parent *frameworkTests
}

func (grp *frameworkSubsetTests) OnlySubtestA(t *testgroup.T) {
	output, sawA, sawB, sawC, sawOthers := grp.testSubtestA(t, true)

	t.Equalf(true, sawA, "should have seen test A, output: %s", output)
	t.Equalf(false, sawB, "should not have seen test B, output: %s", output)
	t.Equalf(false, sawC, "should not have seen test C, output: %s", output)
	t.Equalf(false, sawOthers, "should not have seen other tests, output: %s", output)
}

func (grp *frameworkSubsetTests) EverythingButSubtestA(t *testgroup.T) {
	output, sawA, sawB, sawC, sawOthers := grp.testSubtestA(t, false)

	t.Equalf(false, sawA, "should not have seen test A, output: %s", output)
	t.Equalf(true, sawB, "should have seen test B, output: %s", output)
	t.Equalf(true, sawC, "should have seen test C, output: %s", output)
	t.Equalf(false, sawOthers, "should not have seen other tests, output: %s", output)
}

// Helper routine that either runs ONLY subtestA or everything EXCEPT subtestA.
func (grp *frameworkSubsetTests) testSubtestA(t *testgroup.T, includeA bool) (
	output []byte, sawA, sawB, sawC, sawOthers bool,
) {
	negation := ""
	if !includeA {
		negation = "^"
	}
	runArg := fmt.Sprintf("-run=DocketRunAtTopLevel/[%sA]$", negation)

	output = grp.parent.runGoTest(t, runArg)

	ranTest := regexp.MustCompile(`^=== RUN   Test.+/(Test)?[A-Z]$`)

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
