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

// Command dkt runs docker-compose with a set of docket files.
//
// Usage:
//
//     dkt [OPTIONS] [arguments to docker-compose...]
//
// Options:
//
//     -h, --help            Show this help
//     -v, --version         Show version information
//     -m, --mode=MODE       Set the docket mode (required) [$DOCKET_MODE]
//     -P, --prefix=PREFIX   Set the docket prefix (default: docket) [$DOCKET_PREFIX]
//
// See https://github.com/bloomberg/docket/dkt for more documentation.
//
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/bloomberg/docket/internal/compose"
)

type options struct {
	Mode   string
	Prefix string

	Version bool
	Help    bool
}

func main() {
	envMode := os.Getenv("DOCKET_MODE")
	envPrefix := os.Getenv("DOCKET_PREFIX")

	os.Exit(run(envMode, envPrefix, os.Stdin, os.Stdout, os.Stderr, os.Args[1:]...))
}

func run(envMode, envPrefix string, stdin io.Reader, stdout, stderr io.Writer, args ...string) int {
	opts, remainingArgs, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)

		return 1
	}

	if opts.Mode == "" {
		opts.Mode = envMode
	}
	if opts.Prefix == "" {
		opts.Prefix = envPrefix
	}

	switch {
	case opts.Help:
		return printHelp(stdout)
	case opts.Version:
		return printVersions(stdout)
	case len(remainingArgs) == 0:
		return printHelp(stdout)
	default:
		return runDockerCompose(stdin, stdout, stderr, opts, remainingArgs)
	}
}

type missingParamForOptionError string

func (err missingParamForOptionError) Error() string {
	return fmt.Sprintf("missing parameter to option %s", string(err))
}

func parseArgs(args []string) (options, []string, error) {
	var opts options

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h", arg == "--help":
			opts.Help = true

		case arg == "-v", arg == "--version":
			opts.Version = true

		case arg == "-m", arg == "--mode": // -m NAME or --mode NAME
			if i+1 >= len(args) {
				return opts, nil, missingParamForOptionError(arg)
			}
			opts.Mode = args[i+1]
			i++
		case strings.HasPrefix(arg, "-m"): // -mNAME
			opts.Mode = arg[len("-m"):]
		case strings.HasPrefix(arg, "--mode="): // --mode=NAME
			opts.Mode = arg[len("--mode="):]

		case arg == "-P", arg == "--prefix": // -P NAME or --prefix NAME
			if i+1 >= len(args) {
				return opts, nil, missingParamForOptionError(arg)
			}
			opts.Prefix = args[i+1]
			i++
		case strings.HasPrefix(arg, "-P"): // -PNAME
			opts.Prefix = arg[len("-P"):]
		case strings.HasPrefix(arg, "--prefix="): // --prefix=NAME
			opts.Prefix = arg[len("--prefix="):]

		default:
			return opts, args[i:], nil
		}
	}

	return opts, nil, nil
}

func printHelp(stdout io.Writer) int {
	fmt.Fprint(stdout, `
dkt runs docker-compose with the docker-compose files and generated
configuration that match the given docket mode and prefix.

Any arguments that aren't dkt-specific will be passed through to docker-compose.

Usage:
  dkt [OPTIONS] [arguments to docker-compose...]

Examples:
  dkt config
  dkt up -d
  dkt down

Options:
  -h, --help            Show this help
  -v, --version         Show version information
  -m, --mode=MODE       Set the docket mode (required) [$DOCKET_MODE]
  -P, --prefix=PREFIX   Set the docket prefix (default: docket) [$DOCKET_PREFIX]

Output of 'docker-compose help'
-------------------------------

`)

	runDockerComposeDirectly(nil, stdout, nil, "help")

	return 0
}

func printVersions(stdout io.Writer) int {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		mainMod := buildInfo.Main
		fmt.Fprintf(stdout,
			"dkt from %s %s (%s)\n", mainMod.Path, mainMod.Version, runtime.Version())
	} else {
		fmt.Fprintf(stdout,
			"dkt from github.com/bloomberg/docket in GOPATH (%s)\n", runtime.Version())
	}

	return runDockerComposeDirectly(nil, stdout, nil, "version")
}

func runDockerCompose(
	stdin io.Reader, stdout, stderr io.Writer, opts options, remainingArgs []string,
) int {
	switch remainingArgs[0] {
	case "version":
		if len(remainingArgs) == 1 {
			return printVersions(stdout)
		}

		return runDockerComposeDirectly(stdin, stdout, stderr, remainingArgs...)

	case "help":
		if len(remainingArgs) == 1 {
			return printHelp(stdout)
		}

		return runDockerComposeDirectly(stdin, stdout, stderr, remainingArgs...)

	default:
		return useDocket(stdin, stdout, stderr, opts, remainingArgs)
	}
}

func runDockerComposeDirectly(stdin io.Reader, stdout, stderr io.Writer, args ...string) int {
	cmd := exec.Command("docker-compose", args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(stderr, "failed to run %v: %v\n", cmd.Args, err)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}

		panic(fmt.Sprintf("unanticipated %T: %v", err, err))
	}

	return 0
}

func useDocket(
	stdin io.Reader, stdout, stderr io.Writer, opts options, remainingArgs []string,
) int {
	if opts.Prefix == "" {
		opts.Prefix = "docket"
	}
	if opts.Mode == "" {
		fmt.Fprintf(stderr, "ERROR: use -m|--mode or set $DOCKET_MODE\n")

		return 1
	}

	ctx := context.Background()
	cmp, cleanup, err := compose.NewCompose(ctx, opts.Prefix, opts.Mode)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)

		return 1
	}
	defer func() {
		if err := cleanup(); err != nil {
			fmt.Fprintf(stderr, "ERROR cleaning up: %v\n", err)
		}
	}()

	cmd := cmp.Command(ctx, remainingArgs...)

	// passthrough
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	signal.Ignore(os.Interrupt)

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}

		panic(fmt.Sprintf("unanticipated %T: %v", err, err))
	}

	return 0
}
