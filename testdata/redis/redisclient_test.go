package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bloomberg/docket"
	"gopkg.in/redis.v5"
)

func TestRedisPing(t *testing.T) {
	ctx := context.Background()

	cfgs := docket.ConfigMap{
		"full": {
			ComposeFiles: []string{
				"docker-common.yaml",
				"docker-internal-network.yaml",
			},
			GoTestExec: &docket.GoTestExec{Service: "client"},
		},
		"debug": {
			ComposeFiles: []string{
				"docker-common.yaml",
				"docker-expose-ports.yaml",
			},
		},
	}

	var docketCtx docket.Context

	docket.Run(ctx, cfgs, &docketCtx, t, func() { testRedisPing(t, &docketCtx) })
}

func testRedisPing(t *testing.T, docketCtx *docket.Context) {
	ctx := context.Background()

	var redisAddr string

	if docketCtx.ConfigName() == "debug" {
		defaultRedisPort := 6379
		port, err := docketCtx.ExposedPort(ctx, "redis", defaultRedisPort)
		if err != nil {
			t.Fatalf("could not determine exposed redis port %d; err: %v", defaultRedisPort, err)
		}
		redisAddr = fmt.Sprintf("localhost:%d", port)
	} else {
		redisAddr = os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			t.Fatalf("missing REDIS_ADDR")
		}
	}

	t.Logf("redisAddr = %q\n", redisAddr)
	client := redis.NewClient(&redis.Options{Addr: redisAddr})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	if err != nil {
		t.Fatalf("failed to ping redis: %v", err)
	}
}
