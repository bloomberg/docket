# Hello, world!

The test in [`hello_test.go`](hello_test.go) examines the `HELLO` environment
variable to see if it contains the string `world`.

The test will fail outside the Docker Compose app if the `HELLO` environment
variable is not set.

```console
$ go test
--- FAIL: TestHello (0.00s)
    hello_test.go:17: HELLO had value ""
FAIL
exit status 1
FAIL    .../github.com/bloomberg/docket/testdata/01_hello    0.006s
```

If we set `HELLO=world`, then the test passes.

```console
$ HELLO=world go test
PASS
ok      .../github.com/bloomberg/docket/testdata/01_hello    0.006s
```

The test will pass when run with docket, since [`docket.yaml`](docket.yaml) sets
`HELLO=world` in the `tester` service's environment.

```console
$ DOCKET_MODE=1 DOCKET_DOWN=1 go test
[docket] config [docker-compose --file docket.yaml config]
[docket] up [docker-compose --file docket.yaml --file docket-source-mounts.762772584.yaml up -d]
Creating network "01_hello_default" with the default driver
Creating 01_hello_tester_1 ... done
[docket] up finished
[docket] exec [docker-compose --file docket.yaml --file docket-source-mounts.762772584.yaml exec -T tester go test github.com/bloomberg/docket/testdata/01_hello -run ^TestHello$ -count=1]
ok  	github.com/bloomberg/docket/testdata/01_hello	0.003s
[docket] exec finished
[docket] down [docker-compose --file docket.yaml --file docket-source-mounts.762772584.yaml down]
Stopping 01_hello_tester_1 ... done
Removing 01_hello_tester_1 ... done
Removing network 01_hello_default
[docket] down finished
PASS
ok  	.../github.com/bloomberg/docket/testdata/01_hello	5.189s
```

For more details about how this works and what `DOCKER_MODE` is, see the next
example in [`02_ping-redis`](../02_ping-redis).
