package docket

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

func Test_99_testify_suite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent test suite in short mode")
	}

	runSuiteWithAndWithoutModules(t, &TestifySuiteSuite{
		dir: filepath.Join("testdata", "99_testify-suite"),
	})
}

type TestifySuiteSuite struct {
	gopathOrModulesSuite

	dir string
}

func (s *TestifySuiteSuite) Test_All() {
	s.runGoTest(context.Background())
}

func (s *TestifySuiteSuite) Test_SuiteLevel_OnlySubtestA() {
	output, sawA, sawB, sawC, sawOthers := s.testSubtestA(context.Background(), true)

	s.Equalf(true, sawA, "should have seen TestA, output: %s", output)
	s.Equalf(false, sawB, "should not have seen TestB, output: %s", output)
	s.Equalf(false, sawC, "should not have seen TestC, output: %s", output)
	s.Equalf(false, sawOthers, "should not have seen other tests, output: %s", output)
}

func (s *TestifySuiteSuite) Test_SuiteLevel_EverythingButSubtestA() {
	output, sawA, sawB, sawC, sawOthers := s.testSubtestA(context.Background(), false)

	s.Equalf(false, sawA, "should not have seen TestA, output: %s", output)
	s.Equalf(true, sawB, "should have seen TestB, output: %s", output)
	s.Equalf(true, sawC, "should have seen TestC, output: %s", output)
	s.Equalf(false, sawOthers, "should not have seen other tests, output: %s", output)
}

//------------------------------------------------------------------------------

func (s *TestifySuiteSuite) runGoTest(ctx context.Context, arg ...string) []byte {
	cmd := exec.CommandContext(ctx, "go", "test", "-v")
	cmd.Args = append(cmd.Args, coverageArgs(s.T().Name())...)
	cmd.Args = append(cmd.Args, arg...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), s.GopathEnvOverride()...)
	cmd.Env = append(cmd.Env, "DOCKET_MODE=full", "DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)

	return out
}

// Either run ONLY subtestA or everything EXCEPT subtestA
func (s *TestifySuiteSuite) testSubtestA(ctx context.Context, includeA bool) (output []byte,
	sawA, sawB, sawC, sawOthers bool) {

	negation := ""
	if !includeA {
		negation = "^"
	}
	runArg := fmt.Sprintf("-run=DocketRunAtSuiteLevel/Test[%sA]", negation)

	output = s.runGoTest(ctx, runArg)

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

	return
}
