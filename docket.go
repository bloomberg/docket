package docket

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// Config is a configuration for running tests.
type Config struct {
	// ComposeFiles are the docker-compose configuration files to use.
	// (Think `docker-compose --file FILE`.)
	ComposeFiles []string

	// GoTestExec (optional) specifies how docket should exec `go test` inside a docker container.
	GoTestExec *GoTestExec
}

// GoTestExec specifies how docket should exec `go test` inside a docker container.
type GoTestExec struct {
	// Service is the name of the service (container) in which docket should exec `go test`.
	Service string

	// BuildTags (optional) are build tags to pass to `go test`.
	BuildTags []string
}

// Context tells you information about the current docket test setup.
//
// It is not related to context.Context.
type Context struct {
	activeConfig *namedConfig
}

type namedConfig struct {
	Config
	Name string
}

// ConfigName returns the name of the active Config or a blank string if no Config is being used.
//
// Caveat: When docket execs a test inside a docker container, ConfigName will be empty, since
// the inner test execution is running without an active Config.
func (c Context) ConfigName() string {
	if c.activeConfig == nil {
		return ""
	}

	return c.activeConfig.Name
}

// ExposedPort returns the exposed host port number corresponding to the containerPort for
// a service. If that service does not expose containerPort, it will return an error.
func (c Context) ExposedPort(ctx context.Context, service string, containerPort int) (int, error) {
	if c.activeConfig == nil {
		return -1, fmt.Errorf("no active test config")
	}

	return dockerComposePort(ctx, c.activeConfig.Config, service, containerPort)
}

//----------------------------------------------------------

// ConfigMap is a mapping from a name to a configuration.
type ConfigMap map[string]Config

// Run executes testFunc, possibly using one of the configurations in the ConfigMap (depending
// on whether GO_DOCKET_CONFIG is set), which might mean that it's being executed inside a docker
// container.
//
// If non-nil, testContext will be populated so that it is usable inside testFunc.
func Run(ctx context.Context, configMap ConfigMap, docketCtx *Context, t *testing.T, testFunc func()) {
	configName := os.Getenv("GO_DOCKET_CONFIG")

	if configName == "" {
		testFunc()
		return
	}

	config, ok := configMap[configName]
	if !ok {
		t.Fatalf("no Config found with name %q", configName)
	}

	if docketCtx != nil {
		docketCtx.activeConfig = &namedConfig{
			Name:   configName,
			Config: config,
		}
	}

	if os.Getenv("GO_DOCKET_PULL") != "" {
		if err := dockerComposePull(ctx, config); err != nil {
			t.Fatalf("failed dockerComposePull: %v", err)
		}
	} else {
		// TODO warn about outdated images
	}

	if err := dockerComposeUp(ctx, config); err != nil {
		t.Fatalf("failed dockerComposeUp: %v", err)
	}

	defer func() {
		if os.Getenv("GO_DOCKET_DOWN") != "" {
			if err := dockerComposeDown(ctx, config); err != nil {
				t.Fatalf("failed dockerComposeDown: %v", err)
			}
		} else {
			trace("[docket] leaving docker-compose app running...\n")
		}
	}()

	if config.GoTestExec == nil {
		testFunc()
		return
	}

	if err := dockerComposeExecGoTest(ctx, config, t.Name()); err != nil {
		t.Fatalf("failed exec'ing go test: %v", err)
	}
}
