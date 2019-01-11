# Redis Pinger Service

This Redis Pinger Service uses the `pingredis` module from the previous
[`02_ping-redis`](../02_ping-redis) example and exposes the functionality as an
HTTP service.

This example looks more like a system/integration test: the test makes a real
HTTP request to a real HTTP service.

## Full mode

"Full" mode for this example is straightforward to run -- docket takes care of
everything for you.

```
                                                pinger
                           tester            +-----------+          redis
                        +-----------+        |           |       +---------+
                        |           |        |  go run   |       |         |
go test +----------------> go test +----------> service ----------> redis  |
          (docker exec) |           | (HTTP) |           | (TCP) |         |
                        +-----------+        +-----------+       +---------+
```

Here's what running the test in "full" mode looks like:

```console
$ DOCKET_MODE=full DOCKET_DOWN=1 go test
[docket] config [docker-compose --file docket.yaml --file docket.full.yaml config]
[docket] up [docker-compose --file docket.yaml --file docket.full.yaml --file docket-source-mounts.739064199.yaml up -d]
Creating network "03_redispinger-service_default" with the default driver
Creating 03_redispinger-service_pinger_1 ... done
Creating 03_redispinger-service_tester_1 ... done
Creating 03_redispinger-service_redis_1  ... done
[docket] up finished
[docket] exec [docker-compose --file docket.yaml --file docket.full.yaml --file docket-source-mounts.739064199.yaml exec -T tester go test github.com/bloomberg/docket/testdata/03_redispinger-service -run ^TestService$ -count=1]
ok  	github.com/bloomberg/docket/testdata/03_redispinger-service	0.013s
[docket] exec finished
[docket] down [docker-compose --file docket.yaml --file docket.full.yaml --file docket-source-mounts.739064199.yaml down]
Stopping 03_redispinger-service_redis_1  ... done
Stopping 03_redispinger-service_tester_1 ... done
Stopping 03_redispinger-service_pinger_1 ... done
Removing 03_redispinger-service_redis_1  ... done
Removing 03_redispinger-service_tester_1 ... done
Removing 03_redispinger-service_pinger_1 ... done
Removing network 03_redispinger-service_default
[docket] down finished
PASS
ok  	github.com/bloomberg/docket/testdata/03_redispinger-service	9.696s
```

## Debug mode

"Debug" mode is a bit trickier to get working. Since you might want to run
either the test driver or the service in debugging or profiling, there's a bit
more setup involved.

```
                                       redis
                                    +---------+
                                    |         |
go test +---------> service +--------> redis  |
           (HTTP)             (TCP) |         |
                                    +---------+
```

1.  Since I plan to use "debug" mode for a while, I'll cut down on repetition by
    exporting `DOCKET_MODE` into my environment:

    ```console
    $ export DOCKET_MODE=debug
    ```

2.  I can use the [`dkt`](../../dkt) helper to bring up the dependent container.

    ```console
    $ dkt up -d
    [docket] config [docker-compose --file docket.yaml --file docket.debug.yaml config]
    Running: [docker-compose --file docket.yaml --file docket.debug.yaml up -d]
    Creating network "03_redispinger-service_default" with the default driver
    Creating 03_redispinger-service_redis_1 ... done
    ```

3.  [`docket.debug.yaml`](docket.debug.yaml) configures the `redis` service to
    publish the default Redis port (6379) to the host. Let's find that ephemeral
    host port.

    ```console
    $ dkt port redis 6379
    [docket] config [docker-compose --file docket.yaml --file docket.debug.yaml config]
    Running: [docker-compose --file docket.yaml --file docket.debug.yaml port redis 6379]
    0.0.0.0:32775
    ```

4.  For this example, I'll just run the service with `go run`, but you could
    build and then run the service, or start it in a debugger or with profiling
    enabled. I'll also background it so I can continue with my tests.

    By default, the service will listen on an ephemeral TCP port and print a
    message to stdout telling you which port it's using. I'll add a short
    `sleep` to give that output a chance to show up.

    ```console
    $ go run . & sleep 1
    [1] 34070
    Listening on 127.0.0.1:63416
    ```

5.  I need to set the `REDISPINGER_URL` environment variable to tell the test
    driver where to send the HTTP request.

    ```console
    $ REDISPINGER_URL=http://localhost:63416/?redisAddr=localhost:32775 go test
    [docket] config [docker-compose --file docket.yaml --file docket.debug.yaml config]
    [docket] up [docker-compose --file docket.yaml --file docket.debug.yaml up -d]
    03_redispinger-service_redis_1 is up-to-date
    [docket] up finished
    leaving docker-compose app running...
    PASS
    ok  	.../github.com/bloomberg/docket/testdata/03_redispinger-service	1.147s
    ```

    If I want to run my test really quickly, I can unset `DOCKET_MODE` to skip
    having docket ensure that my Docker Compose app is up.

    ```console
    $ DOCKET_MODE= REDISPINGER_URL="http://localhost:63416/?redisAddr=localhost:32775" go test
    PASS
    ok  	.../github.com/bloomberg/docket/testdata/03_redispinger-service	0.019s
    ```

6.  When I'm done, I'll stop the HTTP service,

    ```console
    $ kill %1
    [1]+  Terminated: 15          go run .
    ```

    clean up the Docker Compose app,

    ```console
    $ dkt down
    [docket] config [docker-compose --file docket.yaml --file docket.debug.yaml config]
    Running: [docker-compose --file docket.yaml --file docket.debug.yaml down]
    Stopping 03_redispinger-service_redis_1 ... done
    Removing 03_redispinger-service_redis_1 ... done
    Removing network 03_redispinger-service_default
    ```

    and clean up my shell environment.

    ```console
    $ unset DOCKET_MODE
    ```
