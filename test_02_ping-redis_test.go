package docket

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Test_02_ping_redis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent test suite in short mode")
	}

	runSuiteWithAndWithoutModules(t, &PingRedisSuite{
		dir: filepath.Join("testdata", "02_ping-redis"),
	})
}

type PingRedisSuite struct {
	gopathOrModulesSuite

	dir string
}

func (s *PingRedisSuite) Test_DebugMode() {
	s.testMode(context.Background(), "debug")
}

func (s *PingRedisSuite) Test_FullMode() {
	s.testMode(context.Background(), "full")
}

//------------------------------------------------------------------------------

func (s *PingRedisSuite) testMode(ctx context.Context, mode string) {
	cmd := exec.CommandContext(ctx, "go", "test", "-v")
	cmd.Args = append(cmd.Args, coverageArgs(s.T().Name())...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), s.GopathEnvOverride()...)
	cmd.Env = append(cmd.Env, "DOCKET_MODE="+mode, "DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)
}
