package main

import (
	"context"
	"os"
	"testing"

	"github.com/bloomberg/docket"
)

func TestEnvVar(t *testing.T) {
	ctx := context.Background()

	cfgs := docket.ConfigMap{
		"full": {
			ComposeFiles: []string{
				"docker-compose.yaml",
			},
			GoTestExec: &docket.GoTestExec{Service: "tester"},
		},
	}

	docket.Run(ctx, cfgs, nil, t, func() { testEnvVar(t) })
}

func testEnvVar(t *testing.T) {
	secret := os.Getenv("DOCKET_SECRET_DATA")
	if secret != "Shh! Don't tell anyone!" {
		t.Errorf("secret had value %q", secret)
	}
}
