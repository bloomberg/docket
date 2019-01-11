# Contributing to docket

## Testing docket

Docket has unit tests as well as integration tests that run the examples in the
`testdata` directory.

### GOPATH and go modules

Docket tries to test itself in both `GOPATH` mode and module-aware mode if
possible.

- Running `go test` inside a `GOPATH` will run both kinds of tests.
  - Use `go test -run /GOPATH/` to run only `GOPATH` mode tests.
  - Use `go test -run /module/` to run only module-aware mode tests.
- Running `go test` outside a `GOPATH` will only run tests in module-aware mode.

### Coverage

To gather coverage, use `-coverprofile` for the main in-process tests and set
`COVERAGE_DIR` to gather coverage from the `go test` child processes. Then use
[`gocovmerge`](https://github.com/wadey/gocovmerge) to merge the coverage data.

```sh
COVERAGE_DIR=COVERAGE go test -v -coverprofile=coverage.root ./... && \
go tool cover -func <(gocovmerge coverage.root $(find COVERAGE -type f))
```
