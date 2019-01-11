package docket

import (
	"bytes"
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
)

func Test_Help(t *testing.T) {
	suite.Run(t, &HelpSuite{})
}

type HelpSuite struct {
	suite.Suite
}

func (s *HelpSuite) Test_runGoTest() {
	cmd := exec.CommandContext(context.Background(), "go", "test", "-help-docket")
	cmd.Args = append(cmd.Args, coverageArgs(s.T().Name())...)

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
