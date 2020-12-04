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

	"github.com/bloomberg/go-testgroup"
)

//------------------------------------------------------------------------------

func Test_dkt(t *testing.T) {
	testgroup.RunSerially(t, new(dktTests)) // cannot parallelize due to chdir
}

type dktTests struct{}

func (grp *dktTests) Help(t *testgroup.T) {
	testcases := [][]string{{"-h"}, {"--help"}, {"help"}, { /* (no args) */ }}

	for _, tc := range testcases {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc), func(t *testgroup.T) {
			var stdout, stderr strings.Builder
			exitCode := run("", "", nil, &stdout, &stderr, tc...)

			t.Zero(exitCode)
			t.Contains(stdout.String(), "dkt")
			t.Contains(stdout.String(), "Usage")
			t.Empty(stderr.String())
		})
	}

	t.Run("compose help for config command", func(t *testgroup.T) {
		var stdout, stderr strings.Builder
		exitCode := run("", "", nil, &stdout, &stderr, "help", "config")

		t.Zero(exitCode)
		t.Contains(stdout.String(), "Usage: config")
		t.Empty(stderr.String())
	})
}

func (grp *dktTests) Version(t *testgroup.T) {
	for _, arg := range []string{"-v", "--version", "version"} {
		arg := arg
		t.Run(arg, func(t *testgroup.T) {
			var stdout, stderr strings.Builder
			exitCode := run("", "", nil, &stdout, &stderr, arg)

			t.Zero(exitCode)
			t.Contains(stdout.String(), "dkt/main from github.com")
			t.Contains(stdout.String(), "docker-compose")
			t.Empty(stderr.String())
		})
	}

	t.Run("version --short", func(t *testgroup.T) {
		var stdout, stderr strings.Builder
		exitCode := run("", "", nil, &stdout, &stderr, "version", "--short")

		t.Zero(exitCode)
		t.NotContains(stdout.String(), "dkt")
		t.Empty(stderr.String())
	})

	t.Run("version bad-arg", func(t *testgroup.T) {
		var stdout, stderr strings.Builder
		exitCode := run("", "", nil, &stdout, &stderr, "version", "bad-arg")

		t.NotZero(exitCode)
		t.Empty(stdout.String())
		t.NotEmpty(stderr.String())
	})
}

func (grp *dktTests) Config(t *testgroup.T) {
	t.Require.NoError(os.Chdir("testdata"))
	defer func() {
		t.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("", "", nil, &stdout, &stderr, "--mode=good", "config")

	t.Zero(exitCode)
	t.Contains(stdout.String(), "version")
	t.Empty(stderr.String())
}

func (grp *dktTests) DocketFailure(t *testgroup.T) {
	t.Require.NoError(os.Chdir("testdata"))
	defer func() {
		t.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("", "", nil, &stdout, &stderr, "--mode=none", "config")

	// There are no docket mode "none" files in this dir, so this should fail.
	t.NotZero(exitCode)
	t.Empty(stdout.String())
	t.Contains(stderr.String(), "ERROR")
}

func (grp *dktTests) ComposeFailure(t *testgroup.T) {
	t.Require.NoError(os.Chdir("testdata"))
	defer func() {
		t.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("", "", nil, &stdout, &stderr, "--mode=good", "dkt_is_the_best")

	t.NotZero(exitCode)
	t.Empty(stdout.String())
	t.Contains(stderr.String(), "No such command")
}

func (grp *dktTests) ModeRequired(t *testgroup.T) {
	var stdout, stderr strings.Builder
	exitCode := run("", "", nil, &stdout, &stderr, "config")

	t.NotZero(exitCode)
	t.Empty(stdout.String())
	t.Contains(stderr.String(), "ERROR")
}

func (grp *dktTests) ModeAndPrefix(t *testgroup.T) {
	t.RunSerially(&modeAndPrefixTests{})
}

//------------------------------------------------------------------------------

type modeAndPrefixTests struct{}

func (grp *modeAndPrefixTests) FromEnv(t *testgroup.T) {
	t.Require.NoError(os.Chdir("testdata"))
	defer func() {
		t.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("good", "docket", nil, &stdout, &stderr, "config")

	t.Zero(exitCode)
	t.Contains(stdout.String(), "version")
	t.Empty(stderr.String())
}

func (grp *modeAndPrefixTests) ArgsOverrideEnv(t *testgroup.T) {
	t.Require.NoError(os.Chdir("testdata"))
	defer func() {
		t.NoError(os.Chdir(".."))
	}()

	var stdout, stderr strings.Builder
	exitCode := run("bad_mode", "bad_prefix", nil, &stdout, &stderr,
		"--mode=good", "--prefix=docket", "config")

	t.Zero(exitCode)
	t.Contains(stdout.String(), "version")
	t.Empty(stderr.String())
}

func (grp *modeAndPrefixTests) MissingArguments(t *testgroup.T) {
	testcases := []string{"-m", "--mode", "-P", "--prefix"}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc, func(t *testgroup.T) {
			var stdout, stderr strings.Builder
			exitCode := run("", "", nil, nil, &stderr, tc)

			t.NotZero(exitCode)
			t.Empty(stdout.String())
			t.Contains(stderr.String(), "missing")
		})
	}
}

func (grp *modeAndPrefixTests) FlagVariations(t *testgroup.T) {
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

		for _, v := range generateFlagVariations(flag.short, flag.long, "VALUE") {
			v := v
			t.Run(fmt.Sprintf("%v", v), func(t *testgroup.T) {
				opts, extras, err := parseArgs(append(v, "extra"))

				t.Equal(flag.expectedOpts, opts)
				t.Equal([]string{"extra"}, extras)
				t.NoError(err)
			})
		}
	}
}

func generateFlagVariations(short, long, value string) [][]string {
	return [][]string{
		{"-" + short, value},
		{"-" + short + value},
		{"--" + long, value},
		{"--" + long + "=" + value},
	}
}
