# Testify Suite

Docket plays nicely with
[testify suites](https://github.com/stretchr/testify/blob/master/README.md#suite-package).
This directory contains examples showing how you can use docket with a testify
suite.

## Wrap an entire suite

`TestDocketRunAtTopLevel` shows how to wrap an entire suite.

The `EntireSuite` struct embeds both a `suite.Suite` (needed by `testify/suite`)
and a `docket.Context`. When I call `docket.Run()`, I pass a pointer to that
part of the struct so docket will fill it in.

When I run the test with docket enabled, docket will run the containing test
entrypoint `TestDocketRunAtTopLevel` inside the `tester` Docker Compose service.
`suite.Run()` will run, which runs the subtest methods attached to the
`EntireSuite` struct.

## Wrap a single function in a suite

`TestDocketRunForSingleSubtest` shows how to wrap a single subtest of a suite.

The `SubtestSuite` struct only needs to embed a `suite.Suite`. The suite has two
subtests:

1. `TestOutsideDocker` does not use docket at all.
2. `TestInsideDocker` wraps its body in a `docket.Run()` call, just as you would
   if you weren't using a testify suite.
