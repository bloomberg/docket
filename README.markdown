# docket

Docket helps you use [Docker Compose](https://docs.docker.com/compose/overview/)
to manage test environments.

## &#x26A0; **_Stability warning: API might change_** &#x26A0;

This pre-1.0.0 API is subject to change as we make improvements.

## Contents

- [Overview](#overview)
- [Examples](#examples)
- [Help](#help)
- [Testing Docket](#testing-docket)
- [Code of Conduct](#code-of-conduct)
- [Contributing](#contributing)
- [License](#license)
- [Security Policy](#security-policy)

## Overview

Docket helps you run a test or test suite inside a multi-container Docker
application using Docker Compose. If requested, docket will run bring up a
Docker Compose app, run the test suite, and optionally shut down the app. If you
don't activate docket, the test will run as if you weren't using docket at all.

Docket is compatible with the standard [`testing`](https://godoc.org/testing)
package (including [`T.Run`](https://godoc.org/testing#T.Run) subtests) as well
as
[`testify/suite`](https://github.com/stretchr/testify/blob/master/README.md#suite-package).

### dkt

Docket includes a command-line utility named [`dkt`](dkt) that helps you run
`docker-compose` commands on your files.

## Examples

You can learn how to use docket by following the examples in the `testdata`
directory.

| `testdata/`                                                                    | Description                                    |
| :----------------------------------------------------------------------------- | :--------------------------------------------- |
| &nbsp;&nbsp;&nbsp; [`01_hello`](testdata/01_hello)                             | Read an environment variable.                  |
| &nbsp;&nbsp;&nbsp; [`02_ping-redis`](testdata/02_ping-redis)                   | Test a function to ping a Redis server         |
| &nbsp;&nbsp;&nbsp; [`03_redispinger-service`](testdata/03_redispinger-service) | Test an HTTP service that pings a Redis server |
| &nbsp;&nbsp;&nbsp; [`99_testify-suite`](testdata/99_testify-suite)             | Use docket with a testify suite.               |

## Help

If your tests import docket, you can run `go test -help-docket` to get help.

### DOCKET_MODE

To enable docket, you'll need to set `DOCKET_MODE` in your environment when you
run `go test`.

For a tutorial on docket modes, see the
[`02_ping-redis`](testdata/02_ping-redis) example.

When you set `DOCKET_MODE=awesome`, docket will look for YAML files (files with
a `.yaml` or `.yml` extension) with names like

- `docket.yaml` (matches any mode)
- `docket.awesome.yaml`
- `docket.awesome.*.yaml`

For more detailed examples, refer to the
[tests](internal/compose/files_test.go).

### Optional

#### DOCKET_DOWN

_Default:_ `false`

If `DOCKET_DOWN` is non-empty, docket will run `docker-compose down` at the end
of each `docket.Run()`.

#### DOCKET_PULL

_Default:_ `false`

If `DOCKET_PULL` is non-empty, docket will run `docker-compose pull` at the
start of each `docket.Run()`.

#### DOCKET_PULL_OPTS

_Default:_ ``

If `DOCKET_PULL_OPTS` is non-empty, docket will add its contents to the
invocation of the `docker-compose pull` command.

For example, to avoid pulling images in parallel, you can set
`DOCKET_PULL_OPTS=--no-parallel` so that docket will run
`docker-compose pull --no-parallel`.

Setting `DOCKET_PULL_OPTS` has no effect if you do not set `DOCKET_PULL=1`.

### Using a custom file prefix

If you need to keep multiple independent docket configurations in the same
directory, you can call `docket.RunPrefix()` to have docket look for YAML files
starting with your custom prefix instead of the default prefix (`"docket"`).

For more detailed examples, refer to the
[tests](internal/compose/files_test.go).

## Testing Docket

Docket has unit tests as well as integration tests that run the examples in the
`testdata` directory.

### Module-aware mode and GOPATH-mode

Docket works in both `GOPATH` mode and module-aware mode, but its tests only run
in one mode at a time. As of Go 1.13, if you haven't overridden `GO111MODULE`,
Go will run in module-aware mode.

To run tests (or other tools) in `GOPATH` mode, you can use the
[`run_in_temp_gopath_with_go_modules_disabled`](run_in_temp_gopath_with_go_modules_disabled)
helper script, which creates a temporary `GOPATH` at `.TEMP_GOPATH`, sets
`GO111MODULE=off`, and runs the script's arguments inside the corresponding
docket package directory (`.TEMP_GOPATH/src/github.com/bloomberg/docket`).

```sh
# module-aware mode, by default
go test ./...

# GOPATH mode (GO111MODULE=off)
./run_in_temp_gopath_with_go_modules_disabled go test ./...
```

### Coverage

To gather coverage, use `-coverprofile` for the main in-process tests and set
`COVERAGE_DIR` to gather coverage from the `go test` child processes. Then, use
[`gocovmerge`](https://github.com/wadey/gocovmerge) to merge the coverage data.

```sh
COVERAGE_DIR=COVERAGE go test -v -coverprofile=coverage.root ./... && \
go tool cover -func <(gocovmerge coverage.root $(find COVERAGE -type f))
```

Note: If you're gathering coverage using
`run_in_temp_gopath_with_go_modules_disabled`, the `COVERAGE_DIR` will be
relative to the temporary docket directory inside `.TEMP_GOPATH` (see above)
unless you give an absolute path as your `COVERAGE_DIR`.

## Code of Conduct

Docket has adopted a
[Code of Conduct](https://github.com/bloomberg/.github/blob/master/CODE_OF_CONDUCT.md).
If you have any concerns about the Code or behavior which you have experienced
in the project, please contact us at opensource@bloomberg.net.

## Contributing

We'd love to hear from you, whether you've found a bug or want to suggest how
docket could be better. Please
[open an issue](https://github.com/bloomberg/docket/issues/new/choose) and let
us know what you think!

If you want to contribute code to the docket project, please be sure to read our
[contribution guidelines](https://github.com/bloomberg/.github/blob/master/CONTRIBUTING.md).
**We highly recommend opening an issue before you start working on your pull
request.** We'd like to talk with you about the change you want to make _before_
you start making it. :smile:

## License

Docket is licensed under the [Apache License, Version 2.0](LICENSE).

## Security Policy

If you believe you have identified a security vulnerability in this project,
please send an email to the project team at opensource@bloomberg.net detailing
the suspected issue and any methods you've found to reproduce it.

Please do _not_ open an issue in the GitHub repository, as we'd prefer to keep
vulnerability reports private until we've had an opportunity to review and
address them. Thank you.
