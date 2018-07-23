# docket

Docket is a library that helps you use
[`docker-compose`][docker-compose-overview] to set up an environment in which
you can run your tests.

&#x26A0; **_Stability warning: API might change_** &#x26A0;

This pre-1.0.0 API is subject to change as we make improvements.

## Overview

Docket helps you run a test or test suite inside a `docker-compose` app. If
requested, docket will run bring up a `docker-compose` app, run the test suite,
and optionally shut down the `docker-compose` app. If you don't activate docket,
the test will run on its own as if you weren't using docket at all.

Docket is compatible with the standard [`testing` package][testing-godoc] as
well as [`testify/suite`][testify-suite-readme].

## Examples

For examples, see the [testdata directory](testdata).

## Running tests with docket

In order to be unobtrusive, docket takes its arguments as environment variables.

For quick help, run `go test -help-docket`, and you'll see this:

```
Help for using docket:

  GO_DOCKET_CONFIG
    To use docket, set this to the name of the config to use.

Optional environment variables:

  GO_DOCKET_DOWN (default off)
      If non-empty, docket will run 'docker-compose down' after each suite.

  GO_DOCKET_PULL (default off)
      If non-empty, docket will run 'docker-compose pull' before each suite.
```

[docker-compose-overview]: https://docs.docker.com/compose/overview/
[testing-godoc]: https://godoc.org/testing
[testify-suite-readme]:
  https://github.com/stretchr/testify/blob/master/README.md#suite-package
