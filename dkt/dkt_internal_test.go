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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

//------------------------------------------------------------------------------

func Test_dkt(t *testing.T) {
	suite.Run(t, new(dktSuite))
}

type dktSuite struct {
	suite.Suite
}

func (s *dktSuite) Test_help() {
	testcases := [][]string{{"-h"}, {"--help"}, {"help"}, { /* (no args) */ }}

	for _, tc := range testcases {
		tc := tc
		s.T().Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			var stdout, stderr strings.Builder
			exitCode := run("", "", nil, &stdout, &stderr, tc...)

			assert := assert.New(t)
			assert.Zero(exitCode)
			assert.Contains(stdout.String(), "dkt")
			assert.Contains(stdout.String(), "Usage")
			assert.Empty(stderr.String())
		})
	}

	s.T().Run("compose help for config command", func(t *testing.T) {
		var stdout, stderr strings.Builder
		exitCode := run("", "", nil, &stdout, &stderr, "help", "config")

		assert := assert.New(t)
		assert.Zero(exitCode)
		assert.Contains(stdout.String(), "Usage: config")
		assert.Empty(stderr.String())
	})
}

func (s *dktSuite) Test_version() {
	for _, arg := range []string{"-v", "--version", "version"} {
		arg := arg
		s.T().Run(arg, func(t *testing.T) {
			var stdout, stderr strings.Builder
			exitCode := run("", "", nil, &stdout, &stderr, arg)

			assert := assert.New(t)
			assert.Zero(exitCode)
			assert.Contains(stdout.String(), "dkt from github.com")
			assert.Contains(stdout.String(), "docker-compose")
			assert.Empty(stderr.String())
		})
	}

	s.T().Run("version --short", func(t *testing.T) {
		var stdout, stderr strings.Builder
		exitCode := run("", "", nil, &stdout, &stderr, "version", "--short")

		assert := assert.New(t)
		assert.Zero(exitCode)
		assert.NotContains(stdout.String(), "dkt")
		assert.Empty(stderr.String())
	})

	s.T().Run("version bad-arg", func(t *testing.T) {
		var stdout, stderr strings.Builder
		exitCode := run("", "", nil, &stdout, &stderr, "version", "bad-arg")

		assert := assert.New(t)
		assert.NotZero(exitCode)
		assert.Empty(stdout.String())
		assert.NotEmpty(stderr.String())
	})
}

func (s *dktSuite) Test_config() {
	s.Require().NoError(os.Chdir("testdata"))
	defer func() {
		s.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("", "", nil, &stdout, &stderr, "--mode=good", "config")

	s.Zero(exitCode)
	s.Contains(stdout.String(), "version")
	s.Empty(stderr.String())
}

func (s *dktSuite) Test_docket_failures() {
	s.Require().NoError(os.Chdir("testdata"))
	defer func() {
		s.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("", "", nil, &stdout, &stderr, "--mode=none", "config")

	// There are no docket mode "none" files in this dir, so this should fail.
	s.NotZero(exitCode)
	s.Empty(stdout.String())
	s.Contains(stderr.String(), "ERROR")
}

func (s *dktSuite) Test_compose_failures() {
	s.Require().NoError(os.Chdir("testdata"))
	defer func() {
		s.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("", "", nil, &stdout, &stderr, "--mode=good", "dkt_is_the_best")

	s.NotZero(exitCode)
	s.Empty(stdout.String())
	s.Contains(stderr.String(), "No such command")
}

func (s *dktSuite) Test_mode_required() {
	var stdout, stderr strings.Builder
	exitCode := run("", "", nil, &stdout, &stderr, "config")

	s.NotZero(exitCode)
	s.Empty(stdout.String())
	s.Contains(stderr.String(), "ERROR")
}

func (s *dktSuite) Test_mode_and_prefix_from_env() {
	s.Require().NoError(os.Chdir("testdata"))
	defer func() {
		s.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("good", "docket", nil, &stdout, &stderr, "config")

	s.Zero(exitCode)
	s.Contains(stdout.String(), "version")
	s.Empty(stderr.String())
}

func (s *dktSuite) Test_mode_and_prefix_args_override_env() {
	s.Require().NoError(os.Chdir("testdata"))
	defer func() {
		s.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("bad_mode", "bad_prefix", nil, &stdout, &stderr,
		"--mode=good", "--prefix=docket", "config")

	s.Zero(exitCode)
	s.Contains(stdout.String(), "version")
	s.Empty(stderr.String())
}

func (s *dktSuite) Test_mode_prefix_missing_argument() {
	testcases := []string{"-m", "--mode", "-P", "--prefix"}

	for _, tc := range testcases {
		tc := tc
		s.T().Run(tc, func(t *testing.T) {
			var stdout, stderr strings.Builder
			exitCode := run("", "", nil, nil, &stderr, tc)
			assert := assert.New(t)
			assert.NotZero(exitCode)
			assert.Empty(stdout.String())
			assert.Contains(stderr.String(), "missing")
		})
	}
}

func (s *dktSuite) Test_parseArgs_mode_and_prefix_variations() {
	flags := []struct {
		short        string
		long         string
		expectedOpts options
	}{
		{
			short:        "m",
			long:         "mode",
			expectedOpts: options{Mode: "VALUE", Prefix: "", Version: false, Help: false},
		},
		{
			short:        "P",
			long:         "prefix",
			expectedOpts: options{Mode: "", Prefix: "VALUE", Version: false, Help: false},
		},
	}

	for _, flag := range flags {
		flag := flag

		for _, v := range flagVariations(flag.short, flag.long, "VALUE") {
			v := v
			s.T().Run(fmt.Sprintf("%v", v), func(t *testing.T) {
				opts, extras, err := parseArgs(append(v, "extra"))

				assert := assert.New(t)
				assert.Equal(flag.expectedOpts, opts)
				assert.Equal([]string{"extra"}, extras)
				assert.NoError(err)
			})
		}
	}
}

func flagVariations(short, long, value string) [][]string {
	return [][]string{
		{"-" + short, value},
		{"-" + short + value},
		{"--" + long, value},
		{"--" + long + "=" + value},
	}
}
