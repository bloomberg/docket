package pingredis

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bloomberg/docket"
)

// TestPingRedis tests the pingRedis function.
func TestPingRedis(t *testing.T) {
	ctx := context.Background()

	var dctx docket.Context

	docket.Run(ctx, &dctx, t, func() {
		testPingRedis(t, &dctx)
	})
}

func testPingRedis(t *testing.T, dctx *docket.Context) {
	ctx := context.Background()

	var redisAddr string

	if dctx.Mode() == "debug" {
		const defaultRedisPort = 6379
		port, err := dctx.PublishedPort(ctx, "redis", defaultRedisPort)
		if err != nil {
			t.Fatalf("could not determine published redis port: %v", err)
		}
		redisAddr = fmt.Sprintf("localhost:%d", port)
	} else {
		redisAddr = os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			t.Fatalf("missing REDIS_ADDR")
		}
	}

	t.Logf("redisAddr = %q", redisAddr)

	pong, err := PingRedis(redisAddr)

	t.Logf("pong = %q", pong)

	if err != nil {
		t.Fatalf("failed to ping redis: %v", err)
	}
}
