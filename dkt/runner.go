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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/debug"

	"github.com/bloomberg/docket/internal/tempbuild"
)

const actualMainPkg = "github.com/bloomberg/docket/dkt/main"

const debugPrefix = "DOCKET_DKT_DEBUG: "

//------------------------------------------------------------------------------

func main() {
	debugTraceEnabled := os.Getenv("DOCKET_DKT_DEBUG") != ""
	keepExecutable := os.Getenv("DOCKET_DKT_KEEP_EXECUTABLE") != ""

	os.Exit(run(debugTraceEnabled, keepExecutable, os.Stdin, os.Stdout, os.Stderr, os.Args[1:]...))
}

func run(
	debugTraceEnabled, keepExecutable bool,
	stdin io.Reader, stdout, stderr io.Writer, args ...string,
) int {
	if debugTraceEnabled {
		printRunnerBuildInfo(debugPrefix, stderr)

		if err := printDiagnostics(debugPrefix, stderr); err != nil {
			fmt.Fprintf(stderr, "ERROR: stopping due to failures when printing diagnostics:\n")
			fmt.Fprintf(stderr, "%v", err)

			return 1
		}
	}

	var singleArg string
	if len(args) == 1 {
		singleArg = args[0]
	}
	switch singleArg {
	case "-v", "--version", "version":
		// Add the runner's version before the dkt and docker-compose versions.
		printRunnerBuildInfo("", stdout)
	}

	if debugTraceEnabled {
		fmt.Fprintf(stderr, "%sbuilding %s\n", debugPrefix, actualMainPkg)
	}
	dktExePath, err := tempbuild.Build(actualMainPkg, "dkt.")
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: failed to build dkt: %v\n", err)
		fmt.Fprintf(stderr, "--- diagnostics follow ---\n")
		_ = printDiagnostics("", stderr)

		return 1
	}
	if !keepExecutable {
		defer os.Remove(dktExePath)
	}

	if debugTraceEnabled {
		fmt.Fprintf(stderr, "%s%v\n", debugPrefix, append([]string{dktExePath}, args...))
	}

	return runDkt(stdin, stdout, stderr, dktExePath, args)
}

func runDkt(stdin io.Reader, stdout, stderr io.Writer, dktExePath string, args []string) int {
	cmd := exec.Command(dktExePath, args...)

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	signal.Ignore(os.Interrupt)

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ProcessState.ExitCode()
		}

		return 1
	}

	return 0
}

//------------------------------------------------------------------------------

func printRunnerBuildInfo(prefix string, w io.Writer) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintf(w, "%sdkt runner from GOPATH (%s)\n", prefix, runtime.Version())

		return
	}

	mainMod := buildInfo.Main
	fmt.Fprintf(w, "%sdkt runner %s (%s)\n", prefix, mainMod.Version, runtime.Version())
}

func printDiagnostics(prefix string, w io.Writer) error {
	curMod, err := goListCurrentModule(prefix, w)
	if err != nil {
		return fmt.Errorf("goListCurrentModule: %w", err)
	}

	if curMod == "" {
		fmt.Fprintf(w, "%snot in module-aware mode\n", prefix)
	} else {
		fmt.Fprintf(w, "%scurrent module: %q\n", prefix, curMod)
	}

	mainPkg, err := goListActualMainPkg(prefix, w)
	if err != nil {
		return fmt.Errorf("goListActualMainPkg: %w", err)
	}

	if mainPkg.Module == nil {
		fmt.Fprintf(w, "%sfound dkt inside the GOPATH: %q\n", prefix, mainPkg.Dir)

		return nil
	}

	mainMod := mainPkg.Module

	var replaceText string
	if replace := mainMod.Replace; replace != nil {
		replaceText = fmt.Sprintf(" Replace={Path=%q Version=%q)", replace.Path, replace.Version)
	}

	fmt.Fprintf(w, "%sfound dkt module: Path=%q Version=%q%s\n",
		prefix, mainMod.Path, mainMod.Version, replaceText)

	return nil
}

type goListInfo struct {
	Dir    string
	Module *moduleInfo
}

type moduleInfo struct {
	Path    string
	Version string
	Replace *moduleInfo
	Main    bool
	Dir     string
}

func goListCurrentModule(prefix string, w io.Writer) (string, error) {
	cmd := exec.Command("go", "list", "-m")
	fmt.Fprintf(w, "%s%v\n", prefix, cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(output, []byte("not using modules")) {
			return "", nil
		}

		return "", fmt.Errorf("unexpected error from go list -m failed %w: %s", err, output)
	}

	return string(bytes.TrimSpace(output)), nil
}

func goListActualMainPkg(prefix string, w io.Writer) (goListInfo, error) {
	cmd := exec.Command("go", "list", "-json", actualMainPkg)
	fmt.Fprintf(w, "%s%v\n", prefix, cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return goListInfo{},
			fmt.Errorf("failed to find actual main package:\n%w\n%s", err, output)
	}

	var mainInfo goListInfo
	if err := json.Unmarshal(output, &mainInfo); err != nil {
		return goListInfo{}, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return mainInfo, nil
}
