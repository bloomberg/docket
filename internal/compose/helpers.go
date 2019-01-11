package compose

import (
	"fmt"
	"path/filepath"
	"strings"
)

// According to `go list`, dirs containing a main aren't in a package, but if we use the name
// relative to a GOPATH, we can use it to invoke `go test <name>`.
func findPackageNameFromCurrentDirAndGOPATH(currentDir string, gopath []string) (string, error) {
	for _, gp := range gopath {
		pathUnderGOPATH, err := filepath.Rel(gp, currentDir)
		if err != nil {
			continue
		}

		srcPrefix := fmt.Sprintf("src%c", filepath.Separator)
		if !strings.HasPrefix(pathUnderGOPATH, srcPrefix) {
			continue
		}

		return filepath.ToSlash(pathUnderGOPATH[len(srcPrefix):]), nil
	}

	return "", fmt.Errorf(
		"could not find package name. currentDir=%q GOPATH=%q", currentDir, gopath)
}

// From go test -h:
//   -run regexp
//       Run only those tests and examples matching the regular expression.
//       For tests, the regular expression is split by unbracketed slash (/)
//       characters into a sequence of regular expressions, and each part
//       of a test's identifier must match the corresponding element in
//       the sequence, if any. Note that possible parents of matches are
//       run too, so that -run=X/Y matches and runs and reports the result
//       of all tests matching X, even those without sub-tests matching Y,
//       because it must run them to look for those sub-tests.
//
// When we run `go test` inside a docker container, we want to re-run this specific test, so we
// use an anchored regexp to exactly match the test and include any other subtest criteria.
//
// testName cannot be empty.
// runArg should be empty if no -run arg was used.
func makeRunArgForTest(testName, runArg string) string {
	if testName == "" {
		panic("testName was empty")
	}

	testParts := strings.Split(testName, "/")
	for i := range testParts {
		testParts[i] = fmt.Sprintf("^%s$", testParts[i])
	}

	var runParts []string
	if runArg != "" {
		runParts = strings.Split(runArg, "/")
	}

	if len(runParts) > len(testParts) {
		return strings.Join(append(testParts, runParts[len(testParts):]...), "/")
	}

	return strings.Join(testParts, "/")
}
