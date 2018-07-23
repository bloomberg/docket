package docket

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestDockerCompose(t *testing.T) {
	var s DockerComposeSuite
	suite.Run(t, &s)
}

type DockerComposeSuite struct {
	suite.Suite
}

func (s *DockerComposeSuite) TestMakeRunArgForTest() {
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
}

func (s *DockerComposeSuite) TestFindPackageNameFromCurrentDirAndGOPATH() {
	goodCases := []struct {
		CurDir, GOPATH, Result string
	}{
		{
			CurDir: filepath.FromSlash("/go/src/package"),
			GOPATH: filepath.FromSlash("/go"),
			Result: "package",
		},
		{
			CurDir: filepath.FromSlash("/go/src/package"),
			GOPATH: strings.Join(
				[]string{filepath.FromSlash("/another"), filepath.FromSlash("/go")},
				string(filepath.ListSeparator)),
			Result: "package",
		},
	}

	for _, c := range goodCases {
		pkg, err := findPackageNameFromCurrentDirAndGOPATH(c.CurDir, c.GOPATH)
		s.Equal(c.Result, pkg, "case: %v", c)
		s.NoError(err, "case: %v", c)
	}

	badCases := []struct {
		CurDir, GOPATH string
	}{
		{"/unrelated/path", "/go"},
		{"/go/src/package", "/mismatched:/gopaths"},
	}

	for _, c := range badCases {
		_, err := findPackageNameFromCurrentDirAndGOPATH(c.CurDir, c.GOPATH)
		s.Error(err, "case: %v", c)
	}
}
