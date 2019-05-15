# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased][]

## [0.3.0][] ([diff][0.3.0-diff]) - 2019-05-15

### Changed

- Docket now sets the working directory of Docker containers that use
  `run go test` or `mount go sources` to be the package's source code directory,
  which more closely matches what `go test` does.

### Fixed

- `dkt` propagates stdin so that `docker-compose` prompts work.

## [0.2.0][] ([diff][0.2.0-diff]) - 2019-01-11

### Added

- Examples in [`testdata`](testdata) are tutorials for how to use docket as well
  as test cases for the test suite, which is fairly exhaustive (over 90%
  coverage).
- Docket automatically mounts your Go sources into services with the right
  labels (`"run go test"` and `"mount go sources"`).
- Docket now supports both `GOPATH` mode and module-aware mode.
- [`dkt`](dkt) is a new tool which wraps `docker-compose` so you can more easily
  interact with your docket setup.

### Changed

- "Configs" have been replaced by "Modes".
  - Use `DOCKET_MODE` instead of `GO_DOCKET_CONFIG`.
  - docket determines which Docker Compose files to use by matching filenames
    based on the mode.
  - `docket.Run` no longer takes a `ConfigMap`.
  - `docket.RunPrefix` allows you to override the default prefix (`docket`).
  - [Labels on Docker Compose services](https://docs.docker.com/compose/compose-file/#labels-2)
    show docket where to bind-mount Go sources and where to run `go test`.
- `docket.Context.ExposedPort()` was renamed to `PublishedPort()`.
- The prefix for environment variables is now `DOCKET_` instead of `GO_DOCKET_`.

## [0.1.0][] - 2018-07-23

First working version of the library.

[unreleased]: https://github.com/bloomberg/docket/compare/v0.3.0...HEAD
[0.3.0-diff]: https://github.com/bloomberg/docket/compare/v0.2.0...v0.3.0
[0.2.0-diff]: https://github.com/bloomberg/docket/compare/v0.1.0...v0.2.0
[0.3.0]: https://github.com/bloomberg/docket/releases/tag/v0.3.0
[0.2.0]: https://github.com/bloomberg/docket/releases/tag/v0.2.0
[0.1.0]: https://github.com/bloomberg/docket/releases/tag/v0.1.0
