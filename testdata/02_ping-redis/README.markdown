# Ping Redis Function

This `pingredis` module provides a simple function that pings a Redis server
address and returns the result.

## Docket modes

This example introduces the concept of a docket **mode**, which is a way to
select between different Docker Compose configurations at the time you run your
tests.

This example supports two modes: "debug" and "full".

When you specify a `DOCKET_MODE`, docket will combine the files whose names
match particular patterns to make a Docker Compose app.

## Shared configuration

The file [`docket.yaml`](docket.yaml) matches any mode, so you can use it to
specify common configurations.

In this example, `docket.yaml` creates a service named `redis` and specifies
which image to use.

## Debug mode

In "debug" mode, the test driver runs outside Docker, and it pings a Redis
instance running in a Docker container (as specified by the Docker Compose
files).

I call this "debug" mode because it makes it easy to run `go test` inside a
debugger.

```
                   redis
                +---------+
                |         |
go test +--------> redis  |
          (TCP) |         |
                +---------+
```

[`docket.yaml`](docket.yaml) already declared a `redis` service, and
[`docket.debug.yaml`](docket.debug.yaml) extends that service definition by
publishing the Redis port so that the test driver can access the Redis instance
from your host.

Instead of picking a fixed host port number for the internal Redis port, I let
Docker choose an ephemeral port for us. After the Docker Compose app has been
started, I can call `PublishedPort()` on our `docket.Context` to discover that
port.

Here's what running the test in "debug" mode looks like:

```console
$ DOCKET_MODE=debug DOCKET_DOWN=1 go test
[docket] config [docker-compose --file docket.yaml --file docket.debug.yaml config]
[docket] up [docker-compose --file docket.yaml --file docket.debug.yaml up -d]
Creating network "02_ping-redis_default" with the default driver
Creating 02_ping-redis_redis_1 ... done
[docket] up finished
[docket] port [docker-compose --file docket.yaml --file docket.debug.yaml port redis 6379]
[docket] down [docker-compose --file docket.yaml --file docket.debug.yaml down]
Stopping 02_ping-redis_redis_1 ... done
Removing 02_ping-redis_redis_1 ... done
Removing network 02_ping-redis_default
[docket] down finished
PASS
ok  	.../github.com/bloomberg/docket/testdata/02_ping-redis	3.574s
```

### Full mode

In "full" mode, the test runs almost entirely inside the Docker Compose app. You
run the first `go test` yourself outside any Docker container, but when docket
realizes that it should run the test(s) inside the `tester` service, it runs
`go test` _again_ with different arguments inside that container.

```
                             tester
                        +---------------+
                        |               |          redis
                        |  wait script  |       +---------+
                        |               |       |         |
go test +------------------> go test +-----------> redis  |
          (docker exec) |               | (TCP) |         |
                        +---------------+       +---------+
```

For this mode, [`docket.full.yaml`](docket.full.yaml) does a few interesting
things:

1. I declare a new `tester` service where docket should run the inner `go test`.
   See the comments in `docket.full.yaml` for a detailed explanation of the
   configuration.

   In particular, it runs a bash script that just waits, keeping the `tester`
   service running so that docket can `docker exec` `go test` into the running
   container.

2. Since the test runs inside Docker, I don't need to publish the Redis
   instance's port to the host. Instead, I mark the default network as being an
   [internal](https://docs.docker.com/compose/compose-file/#internal) network to
   prevent incoming and outgoing network connections. This helps us make sure
   that our test is entirely self-contained as described by our Docker Compose
   files.

Here's what running the test in "full" mode looks like:

```console
$ DOCKET_MODE=full DOCKET_DOWN=1 go test
[docket] config [docker-compose --file docket.yaml --file docket.full.yaml config]
[docket] up [docker-compose --file docket.yaml --file docket.full.yaml --file docket-source-mounts.720177730.yaml up -d]
Creating network "02_ping-redis_default" with the default driver
Creating 02_ping-redis_redis_1  ... done
Creating 02_ping-redis_tester_1 ... done
[docket] up finished
[docket] exec [docker-compose --file docket.yaml --file docket.full.yaml --file docket-source-mounts.720177730.yaml exec -T tester go test github.com/bloomberg/docket/testdata/02_ping-redis -run ^TestPingRedis$ -count=1]
ok  	github.com/bloomberg/docket/testdata/02_ping-redis	0.009s
[docket] exec finished
[docket] down [docker-compose --file docket.yaml --file docket.full.yaml --file docket-source-mounts.720177730.yaml down]
Stopping 02_ping-redis_tester_1 ... done
Stopping 02_ping-redis_redis_1  ... done
Removing 02_ping-redis_tester_1 ... done
Removing 02_ping-redis_redis_1  ... done
Removing network 02_ping-redis_default
[docket] down finished
PASS
ok  	.../github.com/bloomberg/docket/testdata/02_ping-redis	8.444s
```

You can see that docket added another Docker Compose file
`docker-source-mounts.720177730.yaml`, which is a temporary file telling Docker
Compose to mount your Go sources into the docker container so that `go test` can
use them.
