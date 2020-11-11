package docket

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

func Test_help_internal(t *testing.T) {
	suite.Run(t, &HelpInternalSuite{})
}

type HelpInternalSuite struct {
	suite.Suite
}

func (s *HelpInternalSuite) Test_writeHelp() {
	var sb strings.Builder
	writeHelp(&sb)
	s.Regexp("Help for using docket:", sb.String())
}
