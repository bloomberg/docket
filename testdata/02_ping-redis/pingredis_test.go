// Copyright 2019 Bloomberg Finance L.P.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pingredis_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bloomberg/docket"
	pingredis "github.com/bloomberg/docket/testdata/02_ping-redis"
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

	pong, err := pingredis.Ping(redisAddr)

	t.Logf("pong = %q", pong)

	if err != nil {
		t.Fatalf("failed to ping redis: %v", err)
	}

	if pong != "PONG" {
		t.Fatalf(`expected "PONG" but received %q`, pong)
	}
}
