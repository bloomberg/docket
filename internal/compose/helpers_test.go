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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

func Test_Helpers(t *testing.T) {
	suite.Run(t, new(HelpersSuite))
}

type HelpersSuite struct {
	suite.Suite
}

func (s *HelpersSuite) Test_findPackageNameFromDirAndGOPATH() {
	goodCases := []struct {
		PkgDir string
		GOPATH []string
		Result string
	}{
		{
			PkgDir: filepath.FromSlash("/go/src/package"),
			GOPATH: []string{filepath.FromSlash("/go")},
			Result: "package",
		},
		{
			PkgDir: filepath.FromSlash("/go/src/package"),
			GOPATH: []string{filepath.FromSlash("/another"), filepath.FromSlash("/go")},
			Result: "package",
		},
	}

	for _, c := range goodCases {
		pkg, err := findPackageNameFromDirAndGOPATH(c.PkgDir, c.GOPATH)
		s.NoError(err)
		s.Equal(c.Result, pkg, "case: %v", c)
	}

	badCases := []struct {
		PkgDir string
		GOPATH []string
	}{
		{"/unrelated/path", []string{"/go"}},
		{"/go/src/package", []string{"/mismatched", "/gopaths"}},
		{"/absolute/path", []string{"relative/path"}},
	}

	for _, c := range badCases {
		_, err := findPackageNameFromDirAndGOPATH(c.PkgDir, c.GOPATH)
		s.Error(err)
	}
}

func (s *HelpersSuite) Test_makeRunArgForTest() {
	cases := []struct {
		TestName, RunArg, Result string
	}{
		{"top", "", "^top$"},
		{"top", "p", "^top$"},
		{"top", "p/sub", "^top$/sub"},
		{"top/sub", "", "^top$/^sub$"},
		{"top", "/sub", "^top$/sub"},
	}

	for _, c := range cases {
		s.Equal(c.Result, makeRunArgForTest(c.TestName, c.RunArg))
	}

	s.Panics(func() { makeRunArgForTest("", "runArg") })
}
