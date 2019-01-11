// Package docket helps you use Docker Compose to manage test environments.
//
// See the README in https://github.com/bloomberg/docket for usage examples.
//
package docket // import "github.com/bloomberg/docket"

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bloomberg/docket/internal/compose"
)

// Context can tell you information about the active docket environment.
//
// It is not related to context.Context.
type Context struct {
	mode    string
	compose compose.Compose
}

// Mode returns the name of the active mode or a blank string if no mode is being used.
//
// Caveat: When docket runs a test inside a docker container, mode will be empty, since the inner
// test execution is running without an active docket mode.
func (c Context) Mode() string {
	return c.mode
}

// PublishedPort returns the publicly exposed host port number corresponding to the privatePort for
// a service. If that service does not publish privatePort, it will return an error.
func (c Context) PublishedPort(ctx context.Context, service string, privatePort int) (int, error) {
	if c.mode == "" {
		return -1, fmt.Errorf("no active test config")
	}

	return c.compose.GetPort(ctx, service, privatePort)
}

//----------------------------------------------------------

// Run executes testFunc in the proper test environment.
//
// If DOCKET_MODE is set, docket looks for files matching 'docket.yaml', 'docket.MODE.yaml', and
// 'docket.MODE.*.yaml' (.yml files are also allowed). It uses `docker-compose` with the docket
// files (in that order) to set up a test environment, run testFunc, and optionally tear down the
// environment.
//
// If docketCtx is non-nil, it will be populated so that it is usable inside testFunc.
//
// For more documentation and usage examples, see the package's source repository.
func Run(ctx context.Context, docketCtx *Context, t *testing.T, testFunc func()) {
	t.Helper()
	RunPrefix(ctx, docketCtx, t, "docket", testFunc)
}

// RunPrefix acts identically to Run, but it only looks at files starting with prefix.
func RunPrefix(ctx context.Context, docketCtx *Context, t *testing.T, prefix string, testFunc func()) {
	t.Helper()

	mode := os.Getenv("DOCKET_MODE")
	if mode == "" {
		testFunc()
		return
	}

	compose, cleanup, err := compose.NewCompose(ctx, prefix, mode)
	if err != nil {
		t.Fatalf("NewCompose failed: %v", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			t.Fatalf("failed cleanup: %v", err)
		}
	}()

	dctx := Context{
		mode:    mode,
		compose: *compose,
	}

	if docketCtx != nil {
		*docketCtx = dctx
	}

	if os.Getenv("DOCKET_PULL") != "" {
		if err := compose.Pull(ctx); err != nil {
			t.Fatalf("failed compose.Pull: %v", err)
		}
		// } else {
		// TODO warn about outdated images
	}

	if err := compose.Up(ctx); err != nil {
		t.Fatalf("failed compose.Up: %v", err)
	}

	defer func() {
		if os.Getenv("DOCKET_DOWN") != "" {
			if err := compose.Down(ctx); err != nil {
				t.Fatalf("failed compose.Down: %v", err)
			}
		} else {
			fmt.Printf("leaving docker-compose app running...\n") // TODO trace?
		}
	}()

	if err := dctx.compose.RunTestfuncOrExecGoTest(ctx, t.Name(), testFunc); err != nil {
		t.Fatalf("compose.RunTestfuncOrExecGoTest failed: %v", err)
	}
}
