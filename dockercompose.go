package docket

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func dockerComposeUp(ctx context.Context, config Config) error {
	args := append(composeFileArgs(config), "up", "-d")

	cmd := exec.CommandContext(ctx, "docker-compose", args...)

	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GOPATH=%s", determineGOPATH(ctx)),
	)

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	trace("up %v\n", cmd.Args)
	defer trace("up finished\n")

	return cmd.Run()
}

func dockerComposeDown(ctx context.Context, config Config) error {
	args := append(composeFileArgs(config), "down")

	cmd := exec.CommandContext(ctx, "docker-compose", args...)

	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GOPATH=%s", determineGOPATH(ctx)),
	)

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	trace("down %v\n", cmd.Args)
	defer trace("down finished\n")

	return cmd.Run()
}

func dockerComposeExecGoTest(ctx context.Context, config Config, testName string) error {
	var testRunArg string
	if f := flag.Lookup("test.run"); f != nil {
		testRunArg = f.Value.String()
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	testPackage, err := findPackageNameFromCurrentDirAndGOPATH(currentDir, determineGOPATH(ctx))
	if err != nil {
		return fmt.Errorf("failed to find package name: %v", err)
	}

	args := append(
		composeFileArgs(config),
		"exec",
		"-T", // disable pseudo-tty allocation
		config.GoTestExec.Service,
		"go", "test",
		testPackage,
		"-run", makeRunArgForTest(testName, testRunArg))

	if len(config.GoTestExec.BuildTags) > 0 {
		args = append(args, "-tags", strings.Join(config.GoTestExec.BuildTags, " "))
	}

	if testing.Verbose() {
		args = append(args, "-v")
	}

	cmd := exec.CommandContext(ctx, "docker-compose", args...)

	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GOPATH=%s", determineGOPATH(ctx)),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	trace("exec %v\n", cmd.Args)
	defer trace("exec finished\n")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to exec go test: %v", err)
	}

	return nil
}

func dockerComposePort(ctx context.Context, config Config, service string, port int) (int, error) {
	args := append(
		composeFileArgs(config),
		"port",
		service,
		strconv.Itoa(port),
	)

	cmd := exec.CommandContext(ctx, "docker-compose", args...)

	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GOPATH=%s", determineGOPATH(ctx)),
	)

	trace("port %v\n", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("port error: err=%v out=%v", err, out)
	}

	re := regexp.MustCompile(":[[:digit:]]+$")
	match := re.Find(bytes.TrimSpace(out))
	if len(match) == 0 {
		return 0, fmt.Errorf("could not find port number in output: %s", out)
	}

	return strconv.Atoi(string(match[1:])) // drop the leading colon
}

func dockerComposeConfig(ctx context.Context, config Config) ([]byte, error) {
	args := append(composeFileArgs(config), "config")

	cmd := exec.CommandContext(ctx, "docker-compose", args...)

	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GOPATH=%s", determineGOPATH(ctx)),
	)

	trace("config %v\n", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting config: err=%v out=%v", err, out)
	}

	return out, nil
}

func dockerComposePull(ctx context.Context, config Config) error {
	args := append(composeFileArgs(config), "pull")

	cmd := exec.CommandContext(ctx, "docker-compose", args...)

	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GOPATH=%s", determineGOPATH(ctx)),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	trace("pull %v\n", cmd.Args)
	defer trace("pull finished\n")

	return cmd.Run()
}

//----------------------------------------------------------

func determineGOPATH(ctx context.Context) string {
	// Is this weird?
	cmd := exec.CommandContext(ctx, "go", "env", "GOPATH")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("err=%v, out=%s", err, out))
	}
	return strings.TrimSpace(string(out))
}

func composeFileArgs(config Config) []string {
	args := make([]string, 0, len(config.ComposeFiles)*2)

	for _, file := range config.ComposeFiles {
		args = append(args, "-f", file)
	}

	return args
}

// testName cannot be empty. runArg should be empty if no -run arg was used.
func makeRunArgForTest(testName, runArg string) string {
	/*
		-run regexp
			Run only those tests and examples matching the regular expression.
			For tests, the regular expression is split by unbracketed slash (/)
			characters into a sequence of regular expressions, and each part
			of a test's identifier must match the corresponding element in
			the sequence, if any. Note that possible parents of matches are
			run too, so that -run=X/Y matches and runs and reports the result
			of all tests matching X, even those without sub-tests matching Y,
			because it must run them to look for those sub-tests.

		When we run `go test` inside a docker container, we want to re-run this specific test,
		so we use an anchored regexp to exactly match the test and include any other subtest criteria.
	*/

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
	} else {
		return strings.Join(testParts, "/")
	}
}

func findPackageNameFromCurrentDirAndGOPATH(currentDir, gopath string) (string, error) {
	for _, gp := range filepath.SplitList(gopath) {
		pathUnderGOPATH, err := filepath.Rel(gp, currentDir)
		if err != nil {
			continue
		}

		srcPrefix := fmt.Sprintf("src%c", filepath.Separator)
		if !strings.HasPrefix(pathUnderGOPATH, srcPrefix) {
			continue
		}

		return pathUnderGOPATH[len(srcPrefix):], nil
	}

	return "", fmt.Errorf("could not find package name. currentDir=%q GOPATH=%q", currentDir, gopath)
}
