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
