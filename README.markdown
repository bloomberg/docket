# docket

Package docket helps you use [Docker Compose][docker-compose-overview] to manage
test environments.

## &#x26A0; **_Stability warning: API might change_** &#x26A0;

This pre-1.0.0 API is subject to change as we make improvements.

## Overview

Docket helps you run a test or test suite inside a multi-container Docker
application using Docker Compose. If requested, docket will run bring up a
Docker Compose app, run the test suite, and optionally shut down the app. If you
don't activate docket, the test will run as if you weren't using docket at all.

Docket is compatible with the standard [`testing`][testing-godoc] package
(including [`T.Run`][t.run] subtests) as well as
[`testify/suite`][testify-suite-readme].

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

### Using a custom file prefix

If you need to keep multiple independent docket configurations in the same
directory, you can call `docket.RunPrefix()` to have docket look for YAML files
starting with your custom prefix instead of the default prefix (`"docket"`).

For more detailed examples, refer to the
[tests](internal/compose/files_test.go).

## Contributing

See [CONTRIBUTING.markdown](CONTRIBUTING.markdown).

[docker-compose-overview]: https://docs.docker.com/compose/overview/
[testing-godoc]: https://godoc.org/testing
[testify-suite-readme]:
  https://github.com/stretchr/testify/blob/master/README.md#suite-package
[t.run]: https://godoc.org/testing#T.Run
