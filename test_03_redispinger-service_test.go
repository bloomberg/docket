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

package docket_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bloomberg/docket/internal/tempbuild"
	"github.com/bloomberg/go-testgroup"
)

func Test_03_redispinger_service(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent tests in short mode")
	}

	testgroup.RunSerially(t, &RedisPingerTests{
		dir: filepath.Join("testdata", "03_redispinger-service"),
	})
}

type RedisPingerTests struct {
	dir string
}

func (grp *RedisPingerTests) DebugMode(t *testgroup.T) {
	ctx := context.Background()

	dktPath, err := tempbuild.Build(ctx, "github.com/bloomberg/docket/dkt", "dkt.")
	t.Require.NoError(err)
	defer func() { t.NoError(os.Remove(dktPath)) }()

	dkt := func(arg ...string) []byte {
		return grp.runDkt(t, dktPath, append([]string{"--mode=debug"}, arg...)...)
	}

	// Bring up docker compose app and discover redis's port

	dkt("up", "-d")
	defer dkt("down")

	_, redisPort, err := net.SplitHostPort(strings.TrimSpace(string(
		dkt("port", "redis", "6379"))))
	t.Require.NoError(err)

	// Start a pinger service and discover its listener port

	pingerCmd, pingerPort := grp.startPinger(t)
	defer func() {
		t.Require.NoError(pingerCmd.Process.Kill())
		t.Error(pingerCmd.Wait()) // since we killed the process, Wait will return an error
	}()

	// Run go test with REDISPINGER_URL set properly

	testCmd := exec.Command("go", "test", "-v")
	testCmd.Args = append(testCmd.Args, goTestCoverageArgs(t.Name())...)
	testCmd.Args = append(testCmd.Args, goTestRaceDetectorArgs()...)
	testCmd.Dir = grp.dir
	testCmd.Env = append(
		os.Environ(),
		fmt.Sprintf("REDISPINGER_URL=http://localhost:%s/?redisAddr=localhost:%s",
			pingerPort, redisPort),
		"DOCKET_MODE=debug")

	out, err := testCmd.CombinedOutput()
	t.NoError(err, "output: %q", out)
}

func (grp *RedisPingerTests) FullMode(t *testgroup.T) {
	cmd := exec.Command("go", "test", "-v")
	cmd.Args = append(cmd.Args, goTestCoverageArgs(t.Name())...)
	cmd.Args = append(cmd.Args, goTestRaceDetectorArgs()...)
	cmd.Dir = grp.dir
	cmd.Env = append(os.Environ(), "DOCKET_MODE=full", "DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	t.NoError(err, "output: %q", out)
}

//------------------------------------------------------------------------------

func (grp *RedisPingerTests) runDkt(t *testgroup.T, exePath string, arg ...string) []byte {
	cmd := exec.Command(exePath, arg...)
	cmd.Dir = grp.dir

	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Logf("err: %v, stdout: %s, stderr: %s", err, out, exitErr.Stderr)
		} else {
			t.Logf("err: %v, stdout: %s", err, out)
		}
	}
	t.NoError(err)

	return out
}

func (grp *RedisPingerTests) startPinger(t *testgroup.T) (cmd *exec.Cmd, port string) {
	cmd = exec.Command("go", "run", ".")
	cmd.Dir = grp.dir

	stdout, err := cmd.StdoutPipe()
	t.Require.NoError(err)

	t.Require.NoError(cmd.Start())

	scanner := bufio.NewScanner(stdout)
	t.Require.True(scanner.Scan())
	t.Require.NoError(scanner.Err())
	line := scanner.Text()

	// should look like "Listening on 127.0.0.1:1234"
	parts := strings.Fields(line)
	t.Require.Equal(3, len(parts))

	_, port, err = net.SplitHostPort(parts[2])
	t.Require.NoError(err)

	return cmd, port
}
