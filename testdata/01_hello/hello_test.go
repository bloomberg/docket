package main

import (
	"context"
	"os"
	"testing"

	"github.com/bloomberg/docket"
)

func TestHello(t *testing.T) {
	ctx := context.Background()

	docket.Run(ctx, nil, t, func() {
		hello := os.Getenv("HELLO")
		if hello != "world" {
			t.Errorf("HELLO had value %q", hello)
		}
	})
}
