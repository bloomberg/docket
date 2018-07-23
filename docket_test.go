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

	"github.com/stretchr/testify/suite"
)

func TestDocket(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent test suite in short mode")
	}

	suite.Run(t, new(DocketSuite))
}

type DocketSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *DocketSuite) SetupSuite() {
	s.ctx = context.Background()
}

//----------------------------------------------------------

func (s *DocketSuite) TestEnvVarFailsOutsideDocker() {
	cmd := exec.CommandContext(s.ctx, "go", "test", "-v")
	cmd.Dir = filepath.Join("testdata", "envvar")

	// This SHOULD fail, since DOCKET_SECRET_DATA isn't set in the environment.
	out, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("%s", out)
	}
	s.Error(err)
}

func (s *DocketSuite) TestEnvVarSucceedsInsideDocker() {
	cmd := exec.CommandContext(s.ctx, "go", "test", "-v")
	cmd.Dir = filepath.Join("testdata", "envvar")
	cmd.Env = append(os.Environ(), "GO_DOCKET_CONFIG=full", "GO_DOCKET_DOWN=1")

	// Since we activated docket, it should succeed inside our docker-compose app.
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)
}

//----------------------------------------------------------

func testRedis(s *DocketSuite, config string) {
	cmd := exec.CommandContext(s.ctx, "go", "test", "-v")
	cmd.Dir = filepath.Join("testdata", "redis")
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GO_DOCKET_CONFIG=%s", config),
		"GO_DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)
}

func (s *DocketSuite) TestRedisFull() {
	testRedis(s, "full")
}

func (s *DocketSuite) TestRedisDebug() {
	testRedis(s, "debug")
}

//----------------------------------------------------------

func testSuites(s *DocketSuite, testArgs []string) []byte {
	args := append([]string{"test", "-v"}, testArgs...)

	cmd := exec.CommandContext(s.ctx, "go", args...)
	cmd.Dir = filepath.Join("testdata", "suites")
	cmd.Env = append(
		os.Environ(),
		"GO_DOCKET_CONFIG=full",
		"GO_DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)

	return out
}

func (s *DocketSuite) TestSuitesAll() {
	testSuites(s, nil)
}

func testFullSuiteSubtestA(s *DocketSuite, includeA bool) (output []byte, sawA, sawB, sawC, sawOthers bool) {
	negation := ""
	if !includeA {
		negation = "^"
	}
	runArg := fmt.Sprintf("Full/Test[%sA]", negation)

	output = testSuites(s, []string{"-run", runArg})

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

func (s *DocketSuite) TestSuitesFullSuiteOnlySubtestA() {
	output, sawA, sawB, sawC, sawOthers := testFullSuiteSubtestA(s, true)

	s.Equalf(true, sawA, "should have seen TestA, output: %s", output)
	s.Equalf(false, sawB, "should not have seen TestB, output: %s", output)
	s.Equalf(false, sawC, "should not have seen TestC, output: %s", output)
	s.Equalf(false, sawOthers, "should not have seen other tests, output: %s", output)
}

func (s *DocketSuite) TestSuitesFullSuiteExcludingSubtestA() {
	output, sawA, sawB, sawC, sawOthers := testFullSuiteSubtestA(s, false)

	s.Equalf(false, sawA, "should not have seen TestA, output: %s", output)
	s.Equalf(true, sawB, "should have seen TestB, output: %s", output)
	s.Equalf(true, sawC, "should have seen TestC, output: %s", output)
	s.Equalf(false, sawOthers, "should not have seen other tests, output: %s", output)
}
