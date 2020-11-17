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
}

func (s *FilesSuite) Test_filterAndSortFilenames_panics_when_missing_mode_or_prefix() {
	s.Panics(func() { filterAndSortFilenames("", "mode", nil) }, "empty prefix panics")
	s.Panics(func() { filterAndSortFilenames("prefix", "", nil) }, "empty mode panics")
}
