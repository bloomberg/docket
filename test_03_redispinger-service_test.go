package docket

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_03_redispinger_service(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent test suite in short mode")
	}

	runSuiteWithAndWithoutModules(t, &RedisPingerSuite{
		dir: filepath.Join("testdata", "03_redispinger-service"),
	})
}

type RedisPingerSuite struct {
	gopathOrModulesSuite

	dir string
}

func (s *RedisPingerSuite) Test_DebugMode() {
	ctx := context.Background()

	dkt := func(arg ...string) []byte {
		return s.runDkt(ctx, append([]string{"--mode=debug"}, arg...)...)
	}

	// Bring up docker compose app and discover redis's port

	dkt("up", "-d")
	defer dkt("down")

	_, redisPort, err := net.SplitHostPort(strings.TrimSpace(string(
		dkt("port", "redis", "6379"))))
	s.Require().NoError(err)

	// Start a pinger service and discover its listener port

	pingerCmd, pingerPort := s.startPinger(ctx)
	defer func() {
		s.Require().NoError(pingerCmd.Process.Kill())
		pingerCmd.Wait()
	}()

	// Run go test with REDISPINGER_URL set properly

	testCmd := exec.CommandContext(ctx, "go", "test", "-v")
	testCmd.Args = append(testCmd.Args, coverageArgs(s.T().Name())...)
	testCmd.Dir = s.dir
	testCmd.Env = append(os.Environ(), s.GopathEnvOverride()...)
	testCmd.Env = append(
		testCmd.Env,
		fmt.Sprintf("REDISPINGER_URL=http://localhost:%s/?redisAddr=localhost:%s",
			pingerPort, redisPort),
		"DOCKET_MODE=debug")

	out, err := testCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)
}

func (s *RedisPingerSuite) Test_FullMode() {
	cmd := exec.CommandContext(context.Background(), "go", "test", "-v")
	cmd.Args = append(cmd.Args, coverageArgs(s.T().Name())...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), s.GopathEnvOverride()...)
	cmd.Env = append(cmd.Env, "DOCKET_MODE=full", "DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)
}

//------------------------------------------------------------------------------

func (s *RedisPingerSuite) runDkt(ctx context.Context, arg ...string) []byte {
	cmd := exec.CommandContext(ctx, "go", "run", "github.com/bloomberg/docket/dkt")
	cmd.Args = append(cmd.Args, arg...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), s.GopathEnvOverride()...)

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.T().Logf("err: %v, stdout: %s, stderr: %s", err, out, exitErr.Stderr)
		} else {
			s.T().Logf("err: %v, stdout: %s", err, out)
		}
	}
	s.NoError(err)

	return out
}

func (s *RedisPingerSuite) startPinger(ctx context.Context) (cmd *exec.Cmd, port string) {
	cmd = exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), s.GopathEnvOverride()...)

	stdout, err := cmd.StdoutPipe()
	s.Require().NoError(err)

	s.Require().NoError(cmd.Start())

	scanner := bufio.NewScanner(stdout)
	s.Require().True(scanner.Scan())
	s.Require().NoError(scanner.Err())
	line := scanner.Text()

	// should look like "Listening on 127.0.0.1:1234"
	parts := strings.Fields(line)
	s.Require().Equal(3, len(parts))

	_, port, err = net.SplitHostPort(parts[2])
	s.Require().NoError(err)

	return cmd, port
}
