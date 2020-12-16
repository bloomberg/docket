# testgroup

Docket integrates well with
[`github.com/bloomberg/go-testgroup`](https://github.com/bloomberg/go-testgroup).
This directory contains examples showing how you can use docket with a test
group.

## Wrap an entire test group

`TestDocketRunAtTopLevel` shows how to wrap an entire suite.

The `EntireGroup` struct contains a `docket.Context` member. When I call
`docket.Run()`, I pass a pointer to that part of the struct so docket will fill
it in.

When I run the test with docket enabled, docket will run the containing test
entrypoint `TestDocketRunAtTopLevel` inside the `tester` Docker Compose service.
`testgroup.RunSerially()` will run, which runs the subtest methods attached to
the `EntireGroup` struct.

## Wrap a single test inside a group

`TestDocketRunForSingleSubtest` shows how to wrap a single subtest of a rgoup.

The `SubtestGroup` struct has two subtests:

1. `TestOutsideDocker` does not use docket at all.
2. `TestInsideDocker` wraps its body in a `docket.Run()` call, just as you would
   if you weren't using a test group.
