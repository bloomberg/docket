package compose

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

func Test_Files(t *testing.T) {
	suite.Run(t, new(FilesSuite))
}

type FilesSuite struct {
	suite.Suite
}

func (s *FilesSuite) Test_filterAndSortFilenames() {
	cases := []struct {
		prefix string
		mode   string
		files  []string
		result []string
	}{
		{
			prefix: "docket",
			mode:   "mode",
			files: []string{
				"docket.mode.yml",        // good
				"docket.mode.extra.yml",  // good
				"docketyaml",             // BAD: no dot
				"docketyml",              // BAD: no dot
				"docket_yaml",            // BAD: no dot
				"docket_yml",             // BAD: no dot
				"docket.mode_extra.yaml", // BAD: missing dot after mode
				"docket.mode..yaml",      // BAD: missing extra between dots
				"docket.extra.yaml",      // BAD: "extra" doesn't match mode
				"docket.mode.extra.yaml", // good
				"docket.yml",             // good
				"docket.mode.yaml",       // good
				"docket.yaml",            // good
			},
			result: []string{
				"docket.yaml",
				"docket.yml",
				"docket.mode.yaml",
				"docket.mode.yml",
				"docket.mode.extra.yaml",
				"docket.mode.extra.yml",
			},
		},
		{
			prefix: "docket.prefix",
			mode:   "mode",
			files: []string{
				"docket.yaml",                   // BAD: incomplete prefix
				"docket.prefix.extra.yaml",      // BAD: "extra" doesn't match mode
				"docket.prefix_extra.yaml",      // BAD: no dot after prefix
				"docket.prefix.mode.extra.yaml", // good
				"docket.prefix.mode.yaml",       // good
				"docket.prefix.yaml",            // good
			},
			result: []string{
				"docket.prefix.yaml",
				"docket.prefix.mode.yaml",
				"docket.prefix.mode.extra.yaml",
			},
		},
	}

	for _, c := range cases {
		s.Equal(
			c.result,
			filterAndSortFilenames(c.prefix, c.mode, c.files),
			fmt.Sprintf("prefix=%q mode=%q", c.prefix, c.mode))
	}

	s.Panics(func() { filterAndSortFilenames("", "mode", nil) }, "empty prefix panics")
	s.Panics(func() { filterAndSortFilenames("prefix", "", nil) }, "empty mode panics")
}
